package controller

import (
	"fmt"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	dch := NewDomainCheckHealth()
	result := dch.checkHealth(&DomainInfo{
		Domain: "http://18750504314and2.applinzi.com",
	})
	fmt.Println(result)
}
