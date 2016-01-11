package catalog

import (
	"errors"
	"github.com/cloudfoundry-community/aws-broker/common"
	"github.com/jinzhu/gorm"
)

type RDSSettings struct {
	databases map[string]*RDSSetting
}

type RDSSetting struct {
	DB     *gorm.DB
	Config common.DBConfig
}

func InitRDSSettings(secrets *Secrets) (*RDSSettings, error) {
	rdsSettings := RDSSettings{databases: make(map[string]*RDSSetting)}
	for _, rdsSecret := range secrets.RdsSecret.RDSDBSecrets {
		db, err := common.DBInit(&rdsSecret.DBConfig)
		if err == nil {
			rdsSettings.AddRDSSetting(&RDSSetting{DB: db, Config: rdsSecret.DBConfig}, rdsSecret.PlanId)
		} else {
			return nil, err
		}
	}
	return &rdsSettings, nil
}

func (s *RDSSettings) AddRDSSetting(rdsSetting *RDSSetting, planId string) {
	// TODO do additional checks to see if one already exists for that plan id.
	s.databases[planId] = rdsSetting
}

func (s *RDSSettings) GetRDSSettingByPlan(planId string) (*RDSSetting, error) {
	if setting, ok := s.databases[planId]; ok {
		return setting, nil
	}
	return nil, errors.New("Cannot find rds setting by plan id.")
}
