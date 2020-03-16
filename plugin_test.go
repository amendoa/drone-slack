package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/bluele/slack"
	"gotest.tools/assert"
)

func TestNewCommitMessage(t *testing.T) {
	testCases := map[string]struct {
		Msg    string
		Expect Message
	}{
		"Empty Message": {
			Msg:    "",
			Expect: Message{},
		},
		"Title Only": {
			Msg: "title with space",
			Expect: Message{
				msg:   "title with space",
				Title: "title with space",
				Body:  "",
			},
		},
		"Title and Body": {
			Msg: "Title with space\nBody with space",
			Expect: Message{
				msg:   "Title with space\nBody with space",
				Title: "Title with space",
				Body:  "Body with space",
			},
		},
		"Empty Second Line": {
			Msg: "Title with space\n\nBody with space",
			Expect: Message{
				msg:   "Title with space\n\nBody with space",
				Title: "Title with space",
				Body:  "Body with space",
			},
		},
	}

	for name, testCase := range testCases {
		assert.Equal(t, testCase.Expect, newCommitMessage(testCase.Msg), name)
	}
}

func TestDefaultMessage(t *testing.T) {
	repo := getTestRepo()
	build := getTestBuild()

	msg := message(repo, build)
	expectedMessage := "*success* <http://github.com/octocat/hello-world|octocat/hello-world#7fd1a60b> (master) by octocat"

	assert.Equal(t, expectedMessage, msg)
}

func TestDefaultFallbackMessage(t *testing.T) {
	repo := getTestRepo()
	build := getTestBuild()

	msg := fallback(repo, build)
	expectedMessage := "success octocat/hello-world#7fd1a60b (master) by octocat"

	assert.Equal(t, expectedMessage, msg)
}

func TestTemplateMessage(t *testing.T) {
	plugin := getTestPlugin()

	msg, err := templateMessage(plugin.Config.Template, plugin)
	assert.NilError(t, err, "should create message by template without error")
	expectedMessage := `Message Template:
Initial commit

Message body
Initial commit
Message body`

	assert.Equal(t, expectedMessage, msg)
}

func TestTemplateFields(t *testing.T) {
	plugin := getTestPlugin()
	attachment := slack.Attachment{}

	fields, err := templateMessage(fmt.Sprintf(`{"Fields": %s}`, plugin.Config.FieldsTemplate), plugin)
	assert.NilError(t, err, "should create fields by template without error")

	expectedFieldsTemplate := `{"Fields": [{"title": "success", "short": true, "value": "master"}]}`
	assert.Equal(t, expectedFieldsTemplate, fields)

	err = json.Unmarshal([]byte(fields), &attachment)
	assert.NilError(t, err, "should parse json without error")
	assert.Check(t, len(attachment.Fields) > 0)
}

func TestTemplateFallbackMessage(t *testing.T) {
	plugin := getTestPlugin()

	msg, err := templateMessage(plugin.Config.Fallback, plugin)
	assert.NilError(t, err, "should create message by template without error")
	expectedMessage := `Message Template Fallback:
Initial commit
master
success`

	assert.Equal(t, expectedMessage, msg)
}

func getTestPlugin() Plugin {
	return Plugin{
		Repo:   getTestRepo(),
		Build:  getTestBuild(),
		Config: getTestConfig(),
	}
}

func getTestRepo() Repo {
	return Repo{
		Owner: "octocat",
		Name:  "hello-world",
	}
}

func getTestBuild() Build {
	author := Author{
		Username: "octocat",
		Name:     "The Octocat",
		Email:    "octocat@github.com",
		Avatar:   "https://avatars0.githubusercontent.com/u/583231?s=460&v=4",
	}

	return Build{
		Tag:      "1.0.0",
		Event:    "push",
		Number:   1,
		Commit:   "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
		Ref:      "",
		Branch:   "master",
		Author:   author,
		Pull:     "",
		Message:  newCommitMessage("Initial commit\n\nMessage body"),
		DeployTo: "",
		Status:   "success",
		Link:     "http://github.com/octocat/hello-world",
		Started:  1546340400, // 2019-01-01 12:00:00
		Created:  1546340400, // 2019-01-01 12:00:00
	}
}

func getTestConfig() Config {
	t := `Message Template:
{{build.message}}
{{build.message.title}}
{{build.message.body}}`

	tf := `Message Template Fallback:
{{build.message.title}}
{{build.branch}}
{{build.status}}`

	ft := `[{"title": "{{build.status}}", "short": true, "value": "{{build.branch}}"}]`

	return Config{
		Template:       t,
		Fallback:       tf,
		FieldsTemplate: ft,
	}
}
