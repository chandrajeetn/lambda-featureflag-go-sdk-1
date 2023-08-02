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
	LocalEvaluationDeploymentKey              = "server-prdcpdFqASTGAUZ7rifW4B9nPKtkhmpx"
)

type variant struct {
	Value   string      `json:"value,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

type UserProperties struct {
	OrgId            string `json:"org_id,omitempty"`
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
		err = fmt.Errorf("unable to create local evaluation client with given config %v with error %s", config, err.Error())
		panic(err)
	}
}

func fetch(flagName string, user UserProperties) variant {
	flagKeys := []string{flagName}
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
