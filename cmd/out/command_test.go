package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/icrowley/fake"
	"github.com/reconquest/hierr-go"
	"github.com/stretchr/testify/assert"

	"github.com/henry40408/concourse-ssh-resource/internal/models"
	"github.com/henry40408/concourse-ssh-resource/pkg/mockio"
)

func TestOutCommand(t *testing.T) {
	var response outResponse

	words := fake.WordsN(3)
	request, err := json.Marshal(&outRequest{
		Params: models.Params{
			Interpreter: "/bin/sh",
			Script:      fmt.Sprintf(`echo "%s"`, words),
		},
		Source: models.Source{
			Host:     "localhost",
			User:     "root",
			Password: "toor",
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	io, err := mockio.NewMockIO(bytes.NewBuffer(request))
	defer io.Cleanup()
	if !assert.NoError(t, err) {
		return
	}

	err = outCommand(io.In, io.Out, io.Err)
	if !assert.NoError(t, err) {
		return
	}

	// test standard output
	io.Out.Seek(0, 0)
	err = json.NewDecoder(io.Out).Decode(&response)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, response.Version.Timestamp)
	assert.Equal(t, 0, len(response.Metadata))

	// test standard error
	io.Err.Seek(0, 0)
	stderrContent, err := ioutil.ReadAll(io.Err)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, fmt.Sprintf("STDOUT: %s\n", words), string(stderrContent))
}

func TestOutCommandWithInterpreter(t *testing.T) {
	var response outResponse

	words := fake.WordsN(3)
	request, err := json.Marshal(&outRequest{
		Params: models.Params{
			Interpreter: "/usr/bin/python3",
			Script:      fmt.Sprintf(`print("%s")`, words),
		},
		Source: models.Source{
			Host:     "localhost",
			User:     "root",
			Password: "toor",
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	io, err := mockio.NewMockIO(bytes.NewBuffer(request))
	defer io.Cleanup()
	if !assert.NoError(t, err) {
		return
	}

	err = outCommand(io.In, io.Out, io.Err)
	if !assert.NoError(t, err) {
		return
	}

	// test standard output
	io.Out.Seek(0, 0)
	err = json.NewDecoder(io.Out).Decode(&response)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, response.Version.Timestamp)
	assert.Equal(t, 0, len(response.Metadata))

	// test standard error
	io.Err.Seek(0, 0)
	stderrContent, err := ioutil.ReadAll(io.Err)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, fmt.Sprintf("STDOUT: %s\n", words), string(stderrContent))
}

func TestOutCommandWithMalformedJSON(t *testing.T) {
	io, err := mockio.NewMockIO(bytes.NewBuffer([]byte("{")))
	defer io.Cleanup()
	if !assert.NoError(t, err) {
		return
	}

	err = outCommand(io.In, io.Out, io.Err)
	herr := err.(hierr.Error)
	assert.Equal(t, herr.GetMessage(), "unable to parse JSON from standard input")
}

func TestOutCommandWithBadConnectionInfo(t *testing.T) {
	request, err := json.Marshal(&outRequest{
		Params: models.Params{
			Interpreter: "/bin/sh",
			Script:      "uptime",
		},
		Source: models.Source{
			Host:     "localhost",
			User:     "root",
			Password: "",
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	io, err := mockio.NewMockIO(bytes.NewBuffer(request))
	defer io.Cleanup()
	if !assert.NoError(t, err) {
		return
	}

	err = outCommand(io.In, io.Out, io.Err)
	herr := err.(hierr.Error)
	assert.Equal(t, herr.GetMessage(), "unable to run SSH command")
}