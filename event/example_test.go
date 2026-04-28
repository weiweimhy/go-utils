package event

import "fmt"

func ExampleNewBus() {
	bus := NewBus()
	defer bus.Close()

	unsubscribe, ok := bus.Subscribe("user.login", func(eventType string, data any) {
		fmt.Printf("%s: %s\n", eventType, data.(string))
	})
	if !ok {
		return
	}
	defer unsubscribe()

	bus.PublishSync("user.login", "alice")

	// Output:
	// user.login: alice
}
