package bitmap

type Bitmap interface {
	Get(i int) bool
	Set(lf, rg int, val bool)
	SetTrue(i int)
	SetFalse(i int)
}
