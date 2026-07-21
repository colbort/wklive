package svc

import "testing"

func TestChatEventHubFanOut(t *testing.T) {
	hub := NewChatEventHub()
	first, unsubscribeFirst := hub.Subscribe("events")
	defer unsubscribeFirst()
	second, unsubscribeSecond := hub.Subscribe("events")
	defer unsubscribeSecond()

	if err := hub.Publish("events", []byte("message")); err != nil {
		t.Fatal(err)
	}
	for index, subscriber := range []<-chan []byte{first, second} {
		if got := string(<-subscriber); got != "message" {
			t.Fatalf("subscriber %d received %q", index, got)
		}
	}
}

func TestChatEventHubDisconnectsSlowSubscriber(t *testing.T) {
	hub := NewChatEventHub()
	subscriber, unsubscribe := hub.Subscribe("events")
	defer unsubscribe()
	for i := 0; i <= chatEventBuffer; i++ {
		if err := hub.Publish("events", []byte("message")); err != nil {
			t.Fatal(err)
		}
	}
	for range subscriber {
	}
}
