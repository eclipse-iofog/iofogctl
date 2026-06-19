package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
)

type tableFunc = func(string, tableChannel)

var (
	routines = []tableFunc{
		getControllerTable,
		getAgentTable,
		getApplicationTable,
		getSystemApplicationTable,
		getVolumeTable,
		getServiceTable,
		getVolumeMountTable,
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

func getSystemApplicationTable(namespace string, tableChan tableChannel) {
	appExe := newSystemApplicationExecutor(namespace)
	if err := appExe.init(); err != nil {
		tableChan <- tableQuery{err: err}
		return
	}
	table := appExe.generateSystemApplicationOutput()
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

func getEdgeResourceTable(namespace string, tableChan tableChannel) {
	table, err := generateEdgeResourceOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}

func getServiceTable(namespace string, tableChan tableChannel) {
	table, err := generateServicesOutput(namespace)
	if err != nil {
		tableChan <- tableQuery{err: err}
		return
	}
	tableChan <- tableQuery{
		table: table,
	}
}

func getVolumeMountTable(namespace string, tableChan tableChannel) {
	table, err := generateVolumeMountsOutput(namespace)
	tableChan <- tableQuery{
		table: table,
		err:   err,
	}
}
