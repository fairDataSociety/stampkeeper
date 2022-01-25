package api

func (h *Handler) List() []interface{} {
	return h.stampkeeper.List()
}
