package bloom

import (
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestRedisBitSet_New_Set_Test(t *testing.T) {
	store := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       5,
	})
	bitSet := newRedisBitSet(store, "test_key", 1024)
	isSetBefore, err := bitSet.check(nil, []uint{515})
	if err != nil {
		t.Fatal("ss", err)
	}
	t.Log(isSetBefore)

	err = bitSet.set(nil, []uint{515})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRedisBitSet_Add(t *testing.T) {
	store := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       5,
	})
	filter:=New(store, "test_key", 64)
	t.Log([]byte("hello"))
	filter.Add(nil, []byte("hello"))
	t.Log(filter.Exists(nil, []byte("hello")))
}
