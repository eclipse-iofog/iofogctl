/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package install

import (
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
)

type IofogUser struct {
	Name         string
	Surname      string
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
}

type Auth struct {
	URL              string
	Realm            string
	SSL              string
	RealmKey         string
	ControllerClient string
	ControllerSecret string
	ViewerClient     string
}

type Database struct {
	Provider     string
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSL          *bool
	CA           *string
}

type Events struct {
	AuditEnabled     *bool
	RetentionDays    int
	CleanupInterval  int
	CaptureIpAddress *bool
}

// VaultConfig is used by the remote controller install to pass VAULT_* env vars (operator-compatible shape).
type VaultConfig struct {
	Enabled   *bool
	Provider  string
	BasePath  string
	Hashicorp *VaultHashicorpConfig
	Aws       *VaultAwsConfig
	Azure     *VaultAzureConfig
	Google    *VaultGoogleConfig
}

type VaultHashicorpConfig struct {
	Address string
	Token   string
	Mount   string
}

type VaultAwsConfig struct {
	Region      string
	AccessKeyId string
	AccessKey   string
}

type VaultAzureConfig struct {
	URL          string
	TenantId     string
	ClientId     string
	ClientSecret string
}

type VaultGoogleConfig struct {
	ProjectId   string
	Credentials string
}

type Pod struct {
	Name   string
	Status string
}

type K8SControllerConfig struct {
	// User          IofogUser
	Replicas      int32
	ReplicasNats  int32
	Database      Database
	PidBaseDir    string
	EcnViewerPort int
	EcnViewerURL  string
	LogLevel      string
	Auth          Auth
	Events        Events
	Https         *bool
	SecretName    string
	Nats          *cpv3.Nats
	Vault         *cpv3.Vault
}
