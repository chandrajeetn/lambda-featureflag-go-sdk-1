package localEvaluation

import (
	"fmt"
	_ "github.com/chandrajeetn/lambda-flag-go-server/internal/evaluation/lib/linuxArm64"
	_ "github.com/chandrajeetn/lambda-flag-go-server/internal/evaluation/lib/linuxX64"
	_ "github.com/chandrajeetn/lambda-flag-go-server/internal/evaluation/lib/macosArm64"
	_ "github.com/chandrajeetn/lambda-flag-go-server/internal/evaluation/lib/macosX64"
	"github.com/chandrajeetn/lambda-flag-go-server/pkg/experiment"
	"github.com/chandrajeetn/lambda-flag-go-server/pkg/experiment/local"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

var (
	client                                    *local.Client
	LocalEvaluationConfigDebug                = false
	LocalEvaluationConfigServerUrl            = "https://api.lab.amplitude.com/"
	LocalEvaluationConfigPollInterval         = 30
	LocalEvaluationConfigPollerRequestTimeout = 10
	LocalEvaluationDeploymentKey              = ""
)

type variant struct {
	Value   string      `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type UserProperties struct {
	OrgId              string `json:"org_id,omitempty"`
	OrgName            string `json:"org_name,omitempty"`
	OrgStatus          string `json:"org_status,omitempty"`
	Username           string `json:"username,omitempty"`
	Email              string `json:"email,omitempty"`
	Plan               string `json:"plan,omitempty"`
	SubscriptionType   string `json:"subscription_type,omitempty"`
	SubscriptionStatus string `json:"subscription_status,omitempty"`
	HubRegion          string `json:"hub_region,omitempty"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("No .env file found")
	} else {
		fmt.Printf(".env file loaded")
	}

	if os.Getenv("LOCAL_EVALUATION_CONFIG_DEBUG") != "" {
		LocalEvaluationConfigDebug, _ = strconv.ParseBool(os.Getenv("LOCAL_EVALUATION_CONFIG_DEBUG"))
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_SERVER_URL") != "" {
		LocalEvaluationConfigServerUrl = os.Getenv("LOCAL_EVALUATION_CONFIG_SERVER_URL")
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_POLL_INTERVAL") != "" {
		LocalEvaluationConfigPollInterval, _ = strconv.Atoi(os.Getenv("LOCAL_EVALUATION_CONFIG_POLL_INTERVAL"))
	}
	if os.Getenv("LOCAL_EVALUATION_CONFIG_POLLER_REQUEST_TIMEOUT") != "" {
		LocalEvaluationConfigPollerRequestTimeout, _ = strconv.Atoi(os.Getenv("LOCAL_EVALUATION_CONFIG_POLLER_REQUEST_TIMEOUT"))
	}
	if os.Getenv("LOCAL_EVALUATION_DEPLOYMENT_KEY") != "" {
		LocalEvaluationDeploymentKey = os.Getenv("LOCAL_EVALUATION_DEPLOYMENT_KEY")
	}
}

func Initialize() error {
	config := local.Config{
		Debug:                          LocalEvaluationConfigDebug,
		ServerUrl:                      LocalEvaluationConfigServerUrl,
		FlagConfigPollerInterval:       time.Duration(LocalEvaluationConfigPollInterval) * time.Second,
		FlagConfigPollerRequestTimeout: time.Duration(LocalEvaluationConfigPollerRequestTimeout) * time.Second,
	}
	client = local.Initialize(LocalEvaluationDeploymentKey, &config)
	err := client.Start()
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %v with error %s", config, err.Error())
		return err
	}
	return nil
}

func fetch(flagName string, user UserProperties) variant {
	flagKeys := []string{flagName}
	userProp := map[string]interface{}{
		"org_id":              user.OrgId,
		"org_name":            user.OrgName,
		"org_status":          user.OrgStatus,
		"username":            user.Username,
		"email":               user.Email,
		"plan":                user.Plan,
		"subscription_type":   user.SubscriptionType,
		"subscription_status": user.SubscriptionStatus,
		"hub_region":          user.HubRegion,
	}

	expUser := experiment.User{
		UserProperties: userProp,
	}

	variants, err := client.Evaluate(&expUser, flagKeys)
	if err != nil {
		return variant{}
	}

	return variant(variants[flagName])
}

func GetFeatureFlagString(flagName string, user UserProperties) string {
	data := fetch(flagName, user)
	return data.Value
}

func GetFeatureFlagBool(flagName string, user UserProperties) bool {
	data := fetch(flagName, user)
	if val, err := strconv.ParseBool(data.Value); err == nil {
		return val
	}
	return false
}

func GetFeatureFlagPayload(flagName string, user UserProperties) map[string]interface{} {
	data := fetch(flagName, user)
	mapData := make(map[string]interface{})
	mapData["value"] = data.Value
	mapData["payload"] = data.Payload
	return mapData
}
