package types

import "testing"

func Test_Sortlist(t *testing.T) {
	sl := newSortList()
	sl.Add(3)
	sl.Add(2)
	sl.Add(1)
	sl.Delete(2)
	t.Log(sl.GetAll())
	t.Log(sl.GetAll())
}
