package catalog

import (
	"github.com/18F/aws-broker/common"
	"gopkg.in/go-playground/validator.v8"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
)

type RDSDBSecret struct {
	common.DBConfig `yaml:",inline" validate:"required,dive,required"`
	PlanId          string `yaml:"plan_id" validate:"required"`
}

type RDSSecret struct {
	ServiceId    string        `yaml:"service_id" validate:"required"`
	RDSDBSecrets []RDSDBSecret `yaml:"plans" validate:"required,dive,required"`
}

type Secrets struct {
	RdsSecret RDSSecret `yaml:"rds" validate:"required,dive,required"`
}

func InitSecrets(path string) *Secrets {
	var secrets Secrets
	secretsFile := filepath.Join(path, "secrets.yml")
	data, err := ioutil.ReadFile(secretsFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(data, &secrets)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config := &validator.Config{TagName: "validate"}

	validate := validator.New(config)
	validateErr := validate.Struct(secrets)
	if validateErr != nil {
		log.Println(validateErr)
		return nil
	}
	return &secrets
}
