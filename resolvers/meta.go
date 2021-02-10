package resolvers

type MetaResolver struct {
	_Count int32
}

func (r *MetaResolver) Count() int32 {
	return r._Count
}
