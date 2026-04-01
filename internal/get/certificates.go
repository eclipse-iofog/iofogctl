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
	"fmt"
	"strconv"

	// "strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type certificateExecutor struct {
	namespace string
}

func newCertificateExecutor(namespace string) *certificateExecutor {
	a := &certificateExecutor{}
	a.namespace = namespace
	return a
}

func (exe *certificateExecutor) Execute() error {
	printNamespace(exe.namespace)
	if err := generateCertificatesOutput(exe.namespace); err != nil {
		return err
	}
	return nil
}

func (exe *certificateExecutor) GetName() string {
	return ""
}

func generateCertificatesOutput(namespace string) error {
	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	certificateList, err := clt.ListCertificates()
	if err != nil {
		return err
	}

	return tabulateCertificates(certificateList.Certificates)
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%02d-%02d-%d", t.Day(), t.Month(), t.Year())
}

func tabulateCertificates(certificates []client.CertificateInfo) error {
	// Generate table and headers
	table := make([][]string, len(certificates)+1)
	headers := []string{
		"CERTIFICATE",
		// "SUBJECT",
		// "HOSTS",
		"IS_CA",
		"VALID_FROM",
		"VALID_TO",
		// "SERIAL_NUMBER",
		"CA_NAME",
		// "CERTIFICATE_CHAIN",
		"DAYS_REMAINING",
		"IS_EXPIRED",
	}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, certificate := range certificates {
		// Handle CAName pointer
		caName := ""
		if certificate.CAName != nil {
			caName = *certificate.CAName
		}

		// // Handle CertificateChain slice
		// certChain := ""
		// if len(certificate.CertificateChain) > 0 {
		// 	certChain = strings.Join(certificate.CertificateChain, ",")
		// }

		row := []string{
			certificate.Name,
			// certificate.Subject,
			// certificate.Hosts,
			strconv.FormatBool(certificate.IsCA),
			// certificate.ValidFrom.Format(time.RFC3339),
			// certificate.ValidTo.Format(time.RFC3339),
			formatDate(certificate.ValidFrom),
			formatDate(certificate.ValidTo),
			// certificate.SerialNumber,
			caName,
			// certChain,
			strconv.Itoa(certificate.DaysRemaining),
			strconv.FormatBool(certificate.IsExpired),
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	return print(table)
}
