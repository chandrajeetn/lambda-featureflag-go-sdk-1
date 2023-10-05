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

type AmplitudeConfig struct {
	Debug                          bool
	ServerUrl                      string
	FlagConfigPollerInterval       time.Duration
	FlagConfigPollerRequestTimeout time.Duration
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

func InitializeWithConfig(conf AmplitudeConfig, deploymentKey string) {
	client = local.Initialize(deploymentKey, (*local.Config)(&conf))
	err := client.Start()
	if err != nil {
		err = fmt.Errorf("unable to create local evaluation client with given config %+v with error %s", conf, err.Error())
		panic(err)
	}
}

func fetch(user UserProperties) (*local.EvaluationResult, error) {
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
		return nil, err
	}
	return result, nil
}

func getValue(flagName string, user UserProperties) local.EvaluationVariant {
	result, _ := fetch(user)
	if *result != nil {
		if value, ok := (*result)[flagName]; ok {
			return value.Variant
		}
	}
	return local.EvaluationVariant{}
}

func getMapOfValue(user UserProperties) map[string]interface{} {
	flags := make(map[string]interface{})
	result, _ := fetch(user)
	if *result != nil {
		for k, v := range *result {
			if v.IsDefaultVariant {
				continue
			}
			flags[k] = v.Variant.Key
		}
	}
	return flags
}

func GetFeatureFlagString(flagName string, user UserProperties) string {
	data := getValue(flagName, user)
	return data.Key
}

func GetFeatureFlagBool(flagName string, user UserProperties) bool {
	data := getValue(flagName, user)
	if val, err := strconv.ParseBool(data.Key); err == nil {
		return val
	}
	return false
}

func GetFeatureFlagPayload(flagName string, user UserProperties) map[string]interface{} {
	data := getValue(flagName, user)
	mapData := make(map[string]interface{})
	mapData["value"] = data.Key
	mapData["payload"] = data.Payload
	return mapData
}

func GetFeatureFlagByOrg(user UserProperties) map[string]interface{} {
	data := getMapOfValue(user)
	return data
}
