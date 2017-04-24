package log

import (
	"os"
	"testing"
)

func TestHourlyRotateAppender(t *testing.T) {
	const filename = "a.log"
	app, err := NewHourlyRotateAppender(filename)
	if err != nil {
		t.Fatalf("new hourly rotate appender error %v", err)
	}

	log := New("t")

	defer func() {
		app.Close()
		os.Remove(filename)
	}()

	log.SetAppender(app)
	log.Errorf("test string : %v", "only for test")
}
