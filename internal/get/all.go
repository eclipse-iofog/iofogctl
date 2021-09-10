/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
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
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
)

type tableFunc = func(string, tableChannel)

var (
	routines = []tableFunc{
		getControllerTable,
		getAgentTable,
		getApplicationTable,
		getVolumeTable,
		getRouteTable,
	}
)

type tableQuery struct {
	table [][]string
	err   error
}
type tableChannel chan tableQuery

type allExecutor struct {
	namespace string
}

func newAllExecutor(namespace string) *allExecutor {
	exe := &allExecutor{}
	exe.namespace = namespace
	return exe
}

func (exe *allExecutor) GetName() string {
	return ""
}

func (exe *allExecutor) Execute() error {
	// Check namespace exists
	_, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}

	// Add edge resource output if supported
	if err := clientutil.IsEdgeResourceCapable(exe.namespace); err == nil {
		// Add Edge Resources between Agent and Application
		routines = append(routines[:2], append([]tableFunc{getEdgeResourceTable}, routines[2:]...)...)
	}
	// Get tables in parallel
	tableChans := make([]tableChannel, len(routines))
	for idx := range tableChans {
		tableChans[idx] = make(tableChannel, 1)
	}
	for idx, routine := range routines {
		go routine(exe.namespace, tableChans[idx])
	}

	// Start Printing
	printNamespace(exe.namespace)
	for idx := range tableChans {
		tableQuery := <-tableChans[idx]
		if tableQuery.err != nil {
			return tableQuery.err
		}
		if err := print(tableQuery.table); err != nil {
			return err
		}
	}

	return nil
}

func getControllerTable(namespace string, tableChan tableChannel) {
	table, err := generateControllerOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}

func getAgentTable(namespace string, tableChan tableChannel) {
	table, err := generateAgentOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}

func getApplicationTable(namespace string, tableChan tableChannel) {
	appExe := newApplicationExecutor(namespace)
	if err := appExe.init(); err != nil {
		tableChan <- tableQuery{err: err}
		return
	}
	table := appExe.generateApplicationOutput()
	tableChan <- tableQuery{
		table: table,
	}
}

func getVolumeTable(namespace string, tableChan tableChannel) {
	table, err := generateVolumeOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}

func getRouteTable(namespace string, tableChan tableChannel) {
	table, err := generateRouteOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}

func getEdgeResourceTable(namespace string, tableChan tableChannel) {
	table, err := generateEdgeResourceOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}
