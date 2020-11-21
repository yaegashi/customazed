package utils

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

func HasEmpty(items ...string) bool {
	for _, item := range items {
		if item == "" {
			return true
		}
	}
	return false
}

func AzureError(err error) *azure.RequestError {
	if dErr, ok := err.(autorest.DetailedError); ok {
		if sErr, ok := dErr.Original.(*azure.RequestError); ok {
			return sErr
		}
	}
	return nil
}
