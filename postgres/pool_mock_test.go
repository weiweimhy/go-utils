package postgres

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
)

type postgresTestServer struct {
	address  string
	listener net.Listener
	closed   chan struct{}
	mu       sync.Mutex
	conn     net.Conn
	stopped  bool
}

func startPostgresTestServer(t *testing.T, handle func(net.Conn)) *postgresTestServer {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for mock PostgreSQL server: %v", err)
	}
	server := &postgresTestServer{
		address:  listener.Addr().String(),
		listener: listener,
		closed:   make(chan struct{}),
	}
	go func() {
		defer close(server.closed)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			_ = conn.Close()
			return
		}
		server.mu.Lock()
		if server.stopped {
			server.mu.Unlock()
			_ = conn.Close()
			return
		}
		server.conn = conn
		server.mu.Unlock()
		defer conn.Close()
		handle(conn)
	}()
	t.Cleanup(func() {
		server.stop(t)
	})
	return server
}

func (server *postgresTestServer) stop(t *testing.T) {
	t.Helper()
	server.mu.Lock()
	server.stopped = true
	conn := server.conn
	server.mu.Unlock()
	_ = server.listener.Close()
	if conn != nil {
		_ = conn.Close()
	}
	select {
	case <-server.closed:
	case <-time.After(3 * time.Second):
		t.Error("PostgreSQL test server did not stop during cleanup")
	}
}

func (server *postgresTestServer) waitForClose(t *testing.T) {
	t.Helper()
	select {
	case <-server.closed:
	case <-time.After(3 * time.Second):
		t.Fatal("PostgreSQL test server did not stop")
	}
}

type mockPostgres struct {
	*postgresTestServer
}

func startMockPostgres(t *testing.T) *mockPostgres {
	return &mockPostgres{postgresTestServer: startPostgresTestServer(t, func(conn net.Conn) {
		serveMockPostgres(conn, false)
	})}
}

func startFailingMockPostgres(t *testing.T) *mockPostgres {
	return &mockPostgres{postgresTestServer: startPostgresTestServer(t, func(conn net.Conn) {
		serveMockPostgres(conn, true)
	})}
}

type stalledPostgres struct {
	*postgresTestServer
	connectionString string
	startup          chan struct{}
}

func startStalledPostgres(t *testing.T) *stalledPostgres {
	startup := make(chan struct{})
	server := startPostgresTestServer(t, func(conn net.Conn) {
		var firstByte [1]byte
		if _, err := io.ReadFull(conn, firstByte[:]); err != nil {
			return
		}
		close(startup)
		_, _ = io.Copy(io.Discard, conn)
	})
	return &stalledPostgres{
		postgresTestServer: server,
		connectionString:   "postgres://user:password@" + server.address + "/db?sslmode=disable",
		startup:            startup,
	}
}

func (server *stalledPostgres) waitForStartup(t *testing.T) {
	t.Helper()
	select {
	case <-server.startup:
	case <-time.After(3 * time.Second):
		t.Fatal("client did not send a PostgreSQL startup packet")
	}
}

func unusedTCPAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for unused address: %v", err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close temporary listener: %v", err)
	}
	return address
}

func serveMockPostgres(conn net.Conn, failPing bool) {
	backend := pgproto3.NewBackend(conn, conn)
	if !receiveStartup(backend, conn) {
		return
	}
	if sendMessages(backend,
		&pgproto3.AuthenticationOk{},
		&pgproto3.ParameterStatus{Name: "server_version", Value: "16.0"},
		&pgproto3.ParameterStatus{Name: "server_encoding", Value: "UTF8"},
		&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UTF8"},
		&pgproto3.ParameterStatus{Name: "DateStyle", Value: "ISO, MDY"},
		&pgproto3.ParameterStatus{Name: "integer_datetimes", Value: "on"},
		&pgproto3.ParameterStatus{Name: "standard_conforming_strings", Value: "on"},
		&pgproto3.BackendKeyData{ProcessID: 1, SecretKey: []byte{2, 3, 4, 5}},
		&pgproto3.ReadyForQuery{TxStatus: 'I'},
	) != nil {
		return
	}

	resultFormat := int16(pgproto3.TextFormat)
	for {
		message, err := backend.Receive()
		if err != nil {
			return
		}
		var response []pgproto3.BackendMessage
		switch message := message.(type) {
		case *pgproto3.Parse:
			response = append(response, &pgproto3.ParseComplete{})
		case *pgproto3.Bind:
			if len(message.ResultFormatCodes) > 0 {
				resultFormat = message.ResultFormatCodes[0]
			}
			response = append(response, &pgproto3.BindComplete{})
		case *pgproto3.Describe:
			format := int16(pgproto3.TextFormat)
			if message.ObjectType == 'S' {
				response = append(response, &pgproto3.ParameterDescription{})
			} else {
				format = resultFormat
			}
			response = append(response, intRowDescription(format))
		case *pgproto3.Execute:
			if failPing {
				response = append(response, &pgproto3.ErrorResponse{
					Severity:            "ERROR",
					SeverityUnlocalized: "ERROR",
					Code:                "XX000",
					Message:             "mock Ping failure",
				})
			} else {
				response = append(response,
					intDataRow(resultFormat),
					&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
				)
			}
		case *pgproto3.Query:
			if failPing {
				response = append(response, &pgproto3.ErrorResponse{
					Severity:            "ERROR",
					SeverityUnlocalized: "ERROR",
					Code:                "XX000",
					Message:             "mock Ping failure",
				})
			} else {
				response = append(response,
					intRowDescription(pgproto3.TextFormat),
					textDataRow("1"),
					&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")},
				)
			}
			response = append(response, &pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Sync:
			response = append(response, &pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Close:
			response = append(response, &pgproto3.CloseComplete{})
		case *pgproto3.Flush:
			if backend.Flush() != nil {
				return
			}
			continue
		case *pgproto3.Terminate:
			return
		default:
			return
		}
		if sendMessages(backend, response...) != nil {
			return
		}
	}
}

func receiveStartup(backend *pgproto3.Backend, conn net.Conn) bool {
	for {
		message, err := backend.ReceiveStartupMessage()
		if err != nil {
			return false
		}
		switch message.(type) {
		case *pgproto3.StartupMessage:
			return true
		case *pgproto3.SSLRequest:
			if _, err := conn.Write([]byte{'N'}); err != nil {
				return false
			}
		default:
			return false
		}
	}
}

func sendMessages(backend *pgproto3.Backend, messages ...pgproto3.BackendMessage) error {
	for _, message := range messages {
		backend.Send(message)
	}
	return backend.Flush()
}

func intRowDescription(format int16) *pgproto3.RowDescription {
	return &pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{{
		Name:         []byte("?column?"),
		DataTypeOID:  23,
		DataTypeSize: 4,
		TypeModifier: -1,
		Format:       format,
	}}}
}

func intDataRow(format int16) *pgproto3.DataRow {
	value := []byte("1")
	if format == pgproto3.BinaryFormat {
		value = []byte{0, 0, 0, 1}
	}
	return &pgproto3.DataRow{Values: [][]byte{value}}
}

func textDataRow(value string) *pgproto3.DataRow {
	return &pgproto3.DataRow{Values: [][]byte{[]byte(value)}}
}
