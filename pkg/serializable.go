package pkg

import "math/rand"

func (t *Table) ProcessInSerializableMode() {
	j := rand.Intn(t.Count())
	i := rand.Intn(t.Count())
	t.LockTable()
	defer t.UnLockTable()
	t.Update(j, t.Select(i%t.Count())+t.Select((i+1)%t.Count())+t.Select((i+2)%t.Count()))
}
