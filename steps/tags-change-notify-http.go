package steps

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	"github.com/bitflow-stream/go-bitflow/script/reg"
	log "github.com/sirupsen/logrus"
)

// Send a simple HTTP request (configurable through Url and Verb) when a sample appears that contains a previously unseen
// values for a configured tag. If NotificationTimeout is > 0, the notification will be sent again, when a tag value
// appears after not being observed for a certain time.
type TagChangeHttpNotifier struct {
	NoopProcessor
	Url                 string
	Verb                string
	Tag                 string
	NotificationTimeout time.Duration

	lastSeen map[string]time.Time
}

func RegisterTagChangeHttpNotifier(b reg.ProcessorRegistry) {
	b.RegisterAnalysisParamsErr("notify-tag-change-http", func(p *bitflow.SamplePipeline, params map[string]string) error {
		var err error
		timeout := reg.DurationParam(params, "timeout", 0, true, &err)
		verb := reg.StrParam(params, "verb", http.MethodGet, true, &err)
		if err == nil {
			urlStr := params["url"]
			_, err = url.Parse(urlStr) // Check for correctness
			if err == nil {
				p.Add(&TagChangeHttpNotifier{
					Url:                 urlStr,
					Tag:                 params["tag"],
					Verb:                verb,
					NotificationTimeout: timeout,
					lastSeen:            make(map[string]time.Time),
				})
			}
		}
		return err
	}, "", reg.RequiredParams("url", "tag"), reg.OptionalParams("verb", "timeout"))
}

func (t *TagChangeHttpNotifier) String() string {
	return fmt.Sprintf("Notify via HTTP %v, about new values of tag '%v': %v", t.Verb, t.Tag, t.Url)
}

func (t *TagChangeHttpNotifier) Sample(sample *bitflow.Sample, header *bitflow.Header) error {
	value := sample.Tag(t.Tag)
	if value != "" {
		t.handleTagValue(value)
	}
	return t.NoopProcessor.Sample(sample, header)
}

func (t *TagChangeHttpNotifier) handleTagValue(val string) {
	last, ok := t.lastSeen[val]
	now := time.Now()
	if !ok || t.NotificationTimeout <= 0 || now.Sub(last) > t.NotificationTimeout {
		if err := t.notifyAboutTagValue(val); err != nil {
			log.Errorf("Failed to notify via HTTP %v that tag %v=%v appeared (%v): %v", t.Verb, t.Tag, val, t.Url, err)
		}
		t.lastSeen[val] = now
	}
}

func (t *TagChangeHttpNotifier) notifyAboutTagValue(val string) error {
	log.Infof("Notifying through HTTP %v that tag %v=%v appeared: %v", t.Verb, t.Tag, val, t.Url)

	req, err := http.NewRequest(t.Verb, t.Url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		var bodyStr string
		if err != nil {
			bodyStr = fmt.Sprintf("Failed to get response body: %v", err)
		} else {
			bodyStr = "Body: " + string(body)
		}
		return fmt.Errorf("Reponse return status code %v. %v", resp.StatusCode, bodyStr)
	}
	return nil
}
