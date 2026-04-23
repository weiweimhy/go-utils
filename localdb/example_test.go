package localdb

func ExampleOpen() {
	db, err := Open("example-data", "cache.db")
	if err != nil {
		return
	}
	defer db.Close()

	_ = db.Set("demo", "hello", []byte("world"))
}
