package local

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/experiment"

	"github.com/LambdaTest/lambda-featureflag-go-sdk/internal/logger"
)

var clients = map[string]*Client{}
var initMutex = sync.Mutex{}

const EXPOSURE_EVENT_TYPE = "$exposure"

type Client struct {
	log    *logger.Log
	apiKey string
	config *Config
	client *http.Client
	poller *poller
	flags  *string
}

type Event struct {
	EventType       string `json:"event_type"`
	UserId          string `json:"user_id"`
	EventProperties struct {
		FlagKey string      `json:"flag_key"`
		Variant interface{} `json:"variant"`
	} `json:"event_properties"`
}

type ExposurePayload struct {
	ApiKey string  `json:"api_key"`
	Events []Event `json:"events"`
}

func Initialize(apiKey string, config *Config) *Client {
	initMutex.Lock()
	client := clients[apiKey]
	if client == nil {
		if apiKey == "" {
			panic("api key must be set")
		}
		config = fillConfigDefaults(config)
		client = &Client{
			log:    logger.New(config.Debug),
			apiKey: apiKey,
			config: config,
			client: &http.Client{},
			poller: newPoller(),
		}
		client.log.Debug("config: %v", *config)
	}
	initMutex.Unlock()
	return client
}

func (c *Client) Start() error {
	result, err := c.doFlags()
	if err != nil {
		return err
	}
	c.flags = result
	c.poller.Poll(c.config.FlagConfigPollerInterval, func() {
		result, err := c.doFlags()
		if err != nil {
			return
		}
		c.flags = result
	})

	return nil
}

func (c *Client) Evaluate(user *experiment.User, flagKeys []string) (map[string]experiment.Variant, error) {
	variants := make(map[string]experiment.Variant)
	if len(*c.flags) == 0 {
		c.log.Debug("evaluate: no flags")
		return variants, nil

	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	c.log.Debug("evaluate:\n\t- user: %v\n\t- rules: %v\n", string(userJson), *c.flags)

	resultJson := evaluation.Evaluate(*c.flags, string(userJson))
	c.log.Debug("evaluate result: %v\n", resultJson)
	go c.exposure(resultJson, user.UserId)
	var interopResult *interopResult
	err = json.Unmarshal([]byte(resultJson), &interopResult)
	if err != nil {
		return nil, err
	}
	if interopResult.Error != nil {
		return nil, fmt.Errorf("evaluation resulted in error: %v", *interopResult.Error)
	}
	result := interopResult.Result
	filter := len(flagKeys) != 0
	for k, v := range *result {
		if v.IsDefaultVariant || (filter && !contains(flagKeys, k)) {
			continue
		}
		variants[k] = experiment.Variant{
			Value:   v.Variant.Key,
			Payload: v.Variant.Payload,
		}
	}
	return variants, nil
}

func (c *Client) Rules() (map[string]interface{}, error) {
	return c.doRules()
}

func (c *Client) doRules() (map[string]interface{}, error) {
	endpoint, err := url.Parse(c.config.ServerUrl)
	if err != nil {
		return nil, err
	}
	endpoint.Path = "sdk/rules"
	endpoint.RawQuery = "eval_mode=local"
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FlagConfigPollerRequestTimeout)
	defer cancel()
	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	c.log.Debug("rules: %v", string(body))
	var rules []map[string]interface{}
	err = json.Unmarshal(body, &rules)
	if err != nil {
		return nil, err
	}
	var result = make(map[string]interface{})
	for _, rule := range rules {
		flagKey := rule["flagKey"]
		result[fmt.Sprintf("%v", flagKey)] = rule
	}
	return result, nil
}

func (c *Client) Flags() (*string, error) {
	return c.doFlags()
}

func (c *Client) doFlags() (*string, error) {
	path := "/sdk/v1/flags"
	ctx, cancel := context.WithTimeout(context.Background(), c.config.FlagConfigPollerRequestTimeout)
	defer cancel()
	req, err := http.NewRequest("GET", c.config.ServerUrl+path, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Amp-Exp-Library", fmt.Sprintf("experiment-go-server/%v", experiment.VERSION))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	flags := string(body)
	c.log.Debug("flags: %v", flags)
	return &flags, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (c *Client) exposure(resultJson string, userId string) {
	parsePayload := map[string]interface{}{}
	err := json.Unmarshal([]byte(resultJson), &parsePayload)
	if err != nil {
		c.log.Error("unable to parse string %s with error %s", resultJson, err.Error())
		return
	}
	payload := ExposurePayload{}
	payload.ApiKey = os.Getenv("ANALYTICS_API_KEY")
	if result, ok := parsePayload["result"].(map[string]interface{}); ok {
		for flagKey, flagValue := range result {
			event := Event{}
			event.EventType = EXPOSURE_EVENT_TYPE
			event.UserId = userId
			event.EventProperties.FlagKey = flagKey
			if flagResult, ok := flagValue.(map[string]interface{}); ok {
				if variant, ok := flagResult["variant"].(map[string]interface{}); ok {
					event.EventProperties.Variant = variant
				}
			}
			payload.Events = append(payload.Events, event)
		}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.log.Error("unable to marsal payload %+v with error %s", payload, err.Error())
		return
	}
	c.log.Debug("exposure payload : %s", string(payloadBytes))

	ctx, cancel := context.WithTimeout(context.Background(), c.config.FlagConfigPollerRequestTimeout)
	defer cancel()
	req, err := http.NewRequest(http.MethodPost, "https://api2.amplitude.com/2/httpapi", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.log.Error("unable to create request with error %s", err.Error())
		return
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	var client = &http.Client{
		Timeout: c.config.FlagConfigPollerRequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        5,
			MaxIdleConnsPerHost: 5,
			DisableKeepAlives:   true,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		c.log.Error("error %s in making call to amplitude server", err.Error())
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Error("error %s while reading response", err.Error())
		return
	}
	c.log.Debug("exposure response: %s", string(body))
}
