package app

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
    application, err := NewApp()
    assert.NoError(t, err)
    assert.NotNil(t, application.DB)
    assert.NotNil(t, application.Router)
    assert.NotNil(t, application.Logger)
    assert.NotNil(t, application.Limiter)
}
