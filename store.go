package srbac

type Store[K any] interface {
	Insert(obj K) (int64, error)
	Update(id int64, obj K) error
	Get(id int64) (K, error)
	GetAll() ([]K, error)
	Delete(id int64) error
}
