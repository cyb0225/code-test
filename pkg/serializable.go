package pkg

var _ Processor = (*SerializableProcess)(nil)

type SerializableProcess struct {
	t *Table
}

func NewSerializableProcess(t *Table) *SerializableProcess {
	return &SerializableProcess{
		t: t,
	}
}

func (s *SerializableProcess) Process(updateId int, selectIds ...int) {
	s.t.LockTable()
	defer s.t.UnLockTable()

	sum := 0
	for _, v := range selectIds {
		sum += s.t.Select(v)
	}
	s.t.Update(updateId, sum)
}
