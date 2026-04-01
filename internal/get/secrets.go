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

package get

import (
	"strconv"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type secretExecutor struct {
	namespace string
}

func newSecretExecutor(namespace string) *secretExecutor {
	a := &secretExecutor{}
	a.namespace = namespace
	return a
}

func (exe *secretExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateSecretsOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *secretExecutor) GetName() string {
	return ""
}

func generateSecretsOutput(namespace string) error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	secretList, err := clt.ListSecrets()
	if err != nil {
		return err
	}

	return tabulateSecrets(secretList.Secrets)
}

func tabulateSecrets(secrets []client.SecretInfo) error {
	// Generate table and headers
	table := make([][]string, len(secrets)+1)
	headers := []string{
		"SECRET",
		"ID",
		"TYPE",
	}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, secret := range secrets {

		row := []string{
			secret.Name,
			strconv.Itoa(secret.ID),
			secret.Type,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	return print(table)
}
