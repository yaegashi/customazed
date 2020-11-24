package azutil

import (
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

func Error(err error) *azure.RequestError {
	if dErr, ok := err.(autorest.DetailedError); ok {
		if sErr, ok := dErr.Original.(*azure.RequestError); ok {
			return sErr
		}
	}
	return nil
}
