package server

import "testing"

func TestHubBroadcastRoutesOnlySubscribedTopic(t *testing.T) {
	h := NewHub()
	quoteSub := h.NewSubscriber(1)
	depthSub := h.NewSubscriber(1)
	defer h.RemoveSubscriber(quoteSub)
	defer h.RemoveSubscriber(depthSub)

	base := ClientMessage{
		CategoryCode: "crypto",
		Symbol:       "ETHUSDT",
		Market:       "BA",
	}
	quote := base
	quote.Topic = TopicQuote
	depth := base
	depth.Topic = TopicDepth

	if err := h.Subscribe(quoteSub, quote); err != nil {
		t.Fatalf("subscribe quote failed: %v", err)
	}
	if err := h.Subscribe(depthSub, depth); err != nil {
		t.Fatalf("subscribe depth failed: %v", err)
	}

	h.Broadcast(quote, map[string]string{"symbol": "ETHUSDT"})

	select {
	case msg := <-quoteSub.C():
		if msg.Topic != TopicQuote || msg.Symbol != "ETHUSDT" {
			t.Fatalf("unexpected quote message: %+v", msg)
		}
	default:
		t.Fatalf("expected quote subscriber to receive quote")
	}

	select {
	case msg := <-depthSub.C():
		t.Fatalf("depth subscriber should not receive quote, got %+v", msg)
	default:
	}
}
