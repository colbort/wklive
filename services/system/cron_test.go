package main

import (
	"testing"

	"github.com/robfig/cron/v3"
)

func TestCron(t *testing.T) {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	_, err := parser.Parse("0 */5 * * * *")
	if err != nil {
		panic(err)
	}
}
