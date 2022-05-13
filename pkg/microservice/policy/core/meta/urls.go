package meta

import (
	_ "embed"

	"sigs.k8s.io/yaml"

	"github.com/koderover/zadig/pkg/tool/log"
	"github.com/koderover/zadig/pkg/types"
)

//go:embed urls.yaml
var urls []byte
var defaultEmbedUrlConfig *ConfigYaml
var configMapUrlConfig *ConfigYaml
var useConfigMapUrls bool

type ConfigYaml struct {
	ExemptionUrls *ExemptionURLs `json:"exemption_urls"`
	Description   string         `json:"description"`
}

type ExemptionURLs struct {
	Public       []*types.PolicyRule `json:"public"`
	SystemAdmin  []*types.PolicyRule `json:"system_admin"`
	ProjectAdmin []*types.PolicyRule `json:"project_admin"`
}

func init() {
	log.Init(&log.Config{
		Level:    "debug",
		NoCaller: true,
	})
	if err := yaml.Unmarshal(urls, &defaultEmbedUrlConfig); err != nil {
		log.DPanic(err)
	}
}

func GetExemptionsUrls() *ExemptionURLs {
	if !useConfigMapUrls || configMapUrlConfig == nil {
		log.Infof("DefaultExemptionsUrls data use configMap")
		return defaultEmbedUrlConfig.ExemptionUrls
	}
	return defaultEmbedUrlConfig.ExemptionUrls
}

func GetDefaultEmbedUrlConfig() *ConfigYaml {
	return defaultEmbedUrlConfig
}

func RefreshConfigMapByte(b []byte) error {
	if err := yaml.Unmarshal(b, &configMapUrlConfig); err != nil {
		log.Errorf("refresh err:%s", err)
		useConfigMapUrls = false
		return err
	}
	useConfigMapUrls = true
	return nil
}
