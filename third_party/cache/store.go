package cache

type Store interface {
	//Add :将obj根据key加入对应的accumulator中
	Add(obj interface{}) error

	//Update :在key对应的accumulator中更新obj
	Update(obj interface{}) error

	//Delete :在key对应的accumulator中删除obj
	Delete(obj interface{}) error

	//List :返回非空accumulator
	List() []interface{}

	//ListKeys :返回和非空accumulator相关联的keys
	ListKeys() []string

	//Get :返回与obj的key对应的accumulator
	Get(obj interface{}) (item interface{}, exists bool, err error)

	//GetByKey :返回与key对应的accumulator
	GetByKey(key string) (item interface{}, exists bool, err error)

	//Replace :将store中的内容用List中的内容代替
	Replace([]interface{}, string) error
}

// KeyFunc :获取obj中的key
type KeyFunc func(obj interface{}) (string, error)
