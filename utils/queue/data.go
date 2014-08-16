package queue

import (
	"fmt"
	"sync"
)

// data belong to whom
type DataBelong int

const (
	_ DataBelong = iota
	BelongTo1p
	BelongTo2p
	BelongToObs
	BelongToAll
)

// data
type data struct {
	belong DataBelong
	data   interface{}
}

func newData(d interface{}, belong DataBelong) data {
	return data{belong: belong, data: d}
}

func (d data) isBelongTo(belong DataBelong) bool { return d.belong == belong }

// datas
type datas struct {
	d  []data
	mu sync.RWMutex
}

func newDatas() *datas {
	return &datas{
		d: make([]data, 0),
	}
}

// length of data
func (d *datas) length() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.d)
}

// get data from index
func (d *datas) getDataFromIndex(index int, belong DataBelong) []interface{} {
	if d == nil {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	res := make([]interface{}, 0)
	for i := index; i < len(d.d); i++ {
		if d.d[i].isBelongTo(belong) {
			res = append(res, d.d[i].data)
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func (d *datas) setData(data interface{}, belong DataBelong) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.d = append(d.d, newData(data, belong))
}

// table datas
type tableDatas struct {
	datas map[int]*datas
	mu    sync.RWMutex
}

func NewTableDatas() *tableDatas { return &tableDatas{datas: make(map[int]*datas)} }

// debug
func (tds tableDatas) PrintForDebug() string {
	tds.mu.RLock()
	defer tds.mu.RUnlock()
	str := "Table datas now contains the following information:\n"
	for tid, tinfo := range tds.datas {
		str += fmt.Sprintf("	tableId: %d -> length of data is %d\n", tid, tinfo.length())
	}
	return str
}

// get table data
func (tds tableDatas) getTableData(tableId int) *datas {
	tds.mu.RLock()
	defer tds.mu.RUnlock()
	return tds.datas[tableId]
}

// is table data exist
func (tds tableDatas) IsTableExist(tableId int) bool { return tds.getTableData(tableId) != nil }

// new table data
func (tds *tableDatas) NewTableData(tableId int) error {
	tds.mu.Lock()
	defer tds.mu.Unlock()
	if _, ok := tds.datas[tableId]; ok {
		return fmt.Errorf("the table %d is already exist in table datas", tableId)
	}
	tds.datas[tableId] = newDatas()
	return nil
}

// delete the table data
func (tds *tableDatas) DeleteTable(tableId int) {
	tds.mu.Lock()
	defer tds.mu.Unlock()
	delete(tds.datas, tableId)
}

// Get data from index for belong
func (tds tableDatas) GetData(tableId, index int, belong DataBelong) []interface{} {
	return tds.getTableData(tableId).getDataFromIndex(index, belong)
}

// set data
func (tds *tableDatas) SetData(tableId int, data interface{}, belong DataBelong) {
	tds.getTableData(tableId).setData(data, belong)
}
