package localEvaluation

import (
	"fmt"
	_ "github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation/lib/linuxArm64"
	_ "github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation/lib/linuxX64"
	_ "github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation/lib/macosArm64"
	_ "github.com/LambdaTest/lambda-featureflag-go-sdk/internal/evaluation/lib/macosX64"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/experiment"
	"github.com/LambdaTest/lambda-featureflag-go-sdk/pkg/experiment/local"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

var (
	client                                    *local.Client
	LocalEvaluationConfigDebug                = true
	LocalEvaluationConfigServerUrl            = "https://api.lambdatest.com"
	LocalEvaluationConfigPollInterval         = 120
	LocalEvaluationConfigPollerRequestTimeout = 10
	LocalEvaluationDeploymentKey              = "server-jAqqJaX3l8PgNiJpcv9j20ywPzANQQFh"
)

type UserProperties struct {
	OrgId            string `json:"org_id,omitempty"`
	UserId           string `json:"user_id,omitempty"`
	OrgName          string `json:"org_name,omitempty"`
	Username         string `json:"username,omitempty"`
	UserStatus       string `json:"user_status,omitempty"`
	Email            string `json:"email,omitempty"`
	Plan             string `json:"plan,omitempty"`
	SubscriptionType string `json:"subscription_type,omitempty"`
	HubRegion        string `json:"hub_region,omitempty"`
	InfraProvider    string `json:"infra_provider,omitempty"`
	TemplateId       string `json:"template_id,omitempty"`
}

type Config struct {
	Debug                          bool
	ServerUrl                      string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
}

type AmplitudeVariant struct {
	Value   string      `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
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

func Initialize() {
	config := local.Config{
		Debug:                          LocalEvaluationConfigDebug,
		ServerUrl:                      LocalEvaluationConfigServerUrl,
		FlagConfigPollerInterval:       time.Duration(LocalEvaluationConfigPollInterval) * time.Second,
		FlagConfigPollerRequestTimeout: time.Duration(LocalEvaluationConfigPollerRequestTimeout) * time.Second,
	}
	client = local.Initialize(LocalEvaluationDeploymentKey, &config)
	err := client.Start()
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %+v with error %s", config, err.Error())
		panic(err)
	}
}

func InitializeWithConfig(conf Config) {
	client = local.Initialize(LocalEvaluationDeploymentKey, (*local.Config)(&conf))
	err := client.Start()
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %+v with error %s", conf, err.Error())
		panic(err)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func fetch(flagKeys []string, user UserProperties) map[string]AmplitudeVariant {
	variants := make(map[string]AmplitudeVariant)
	userProp := map[string]interface{}{
		"org_id":            user.OrgId,
		"org_name":          user.OrgName,
		"username":          user.Username,
		"email":             user.Email,
		"plan":              user.Plan,
		"subscription_type": user.SubscriptionType,
		"hub_region":        user.HubRegion,
		"infra_provider":    user.InfraProvider,
		"template_id":       user.TemplateId,
	}
	expUser := experiment.User{
		UserId:         user.UserId,
		UserProperties: userProp,
	}

	result, err := client.EvaluateByOrg(&expUser)
	if err != nil {
		return variants
	}
	filter := len(flagKeys) != 0
	for k, v := range *result {
		if v.IsDefaultVariant {
			continue
		}
		if !filter {
			variants[k] = AmplitudeVariant{
				Value:   v.Variant.Key,
				Payload: v.Variant.Payload,
			}
			continue
		}
		if contains(flagKeys, k) {
			variants[k] = AmplitudeVariant{
				Value:   v.Variant.Key,
				Payload: v.Variant.Payload,
			}
		}
	}

	return variants
}

func GetFeatureFlagString(flagName string, user UserProperties) string {
	flagKeys := []string{flagName}
	data := fetch(flagKeys, user)
	return data[flagName].Value
}

func GetFeatureFlagBool(flagName string, user UserProperties) bool {
	flagKeys := []string{flagName}
	data := fetch(flagKeys, user)
	if val, err := strconv.ParseBool(data[flagName].Value); err == nil {
		return val
	}
	return false
}

func GetFeatureFlagPayload(flagName string, user UserProperties) map[string]interface{} {
	flagKeys := []string{flagName}
	data := fetch(flagKeys, user)
	mapData := make(map[string]interface{})
	mapData["value"] = data[flagName].Value
	mapData["payload"] = data[flagName].Payload
	return mapData
}

func GetFeatureFlagByOrg(user UserProperties) map[string]AmplitudeVariant {
	flagKeys := []string{}
	data := fetch(flagKeys, user)
	return data
}
