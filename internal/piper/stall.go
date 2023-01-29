package piper

func genTickets(n int) chan struct{} {
	tickets := make(chan struct{}, n)
	n = max(n, 1)
	for i := 0; i < n; i++ {
		tickets <- struct{}{}
	}
	return tickets
}
