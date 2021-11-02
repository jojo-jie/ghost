package bloom

import (
	"context"
	"errors"
	"fmt"
	"ghost/pkg/hash"
	"github.com/go-redis/redis/v8"

	"strconv"
	"time"
)

const (
	// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
	// maps as k in the error rate table
	maps      = 14
	setScript = `
for _, offset in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], offset, 1)
end
`
	testScript = `
for _, offset in ipairs(ARGV) do
	if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
		return false
	end
end
return true
`
)

var ErrTooLargeOffset = errors.New("too large offset")

type (
	// A Filter is a bloom filter.
	Filter struct {
		bits   uint
		bitSet bitSetProvider
	}

	bitSetProvider interface {
		check(context.Context, []uint) (bool, error)
		set(context.Context, []uint) error
	}
)

// New create a Filter, store is the backed redis, key is the key for the bloom filter,
// bits is how many bits will be used, maps is how many hashes for each addition.
// best practices:
// elements - means how many actual elements
// when maps = 14, formula: 0.7*(bits/maps), bits = 20*elements, the error rate is 0.000067 < 1e-4
// for detailed error rate table, see http://pages.cs.wisc.edu/~cao/papers/summary-cache/node8.html
func New(store *redis.Client, key string, bits uint) *Filter {
	return &Filter{
		bits:   bits,
		bitSet: newRedisBitSet(store, key, bits),
	}
}

// Add adds data into f.
func (f *Filter) Add(ctx context.Context, data []byte) error {
	locations := f.getLocations(data)
	return f.bitSet.set(ctx, locations)
}

// Exists checks if data is in f.
func (f *Filter) Exists(ctx context.Context, data []byte) (bool, error) {
	locations := f.getLocations(data)
	fmt.Println("locations", locations)
	isSet, err := f.bitSet.check(ctx, locations)
	if err != nil {
		return false, err
	}
	if !isSet {
		return false, nil
	}

	return true, nil
}

func (f *Filter) getLocations(data []byte) []uint {
	locations := make([]uint, maps)
	for i := uint(0); i < maps; i++ {
		hashValue := hash.Hash(append(data, byte(i)))
		locations[i] = uint(hashValue % uint64(f.bits))
	}

	return locations
}

type redisBitSet struct {
	store *redis.Client
	key   string
	bits  uint
}

func newRedisBitSet(store *redis.Client, key string, bits uint) *redisBitSet {
	return &redisBitSet{
		store: store,
		key:   key,
		bits:  bits,
	}
}

// buildOffsetArgs 异常检查转字符串
func (r *redisBitSet) buildOffsetArgs(offsets []uint) ([]string, error) {
	var args []string
	for _, offset := range offsets {
		//因为前面对bits求模了，所以正常情况这里是不会走到的
		if offset >= r.bits {
			return nil, ErrTooLargeOffset
		}
		// uint数字转字符
		args = append(args, strconv.FormatUint(uint64(offset), 10))
	}
	return args, nil
}

func (r *redisBitSet) check(ctx context.Context, offsets []uint) (bool, error) {
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return false, err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	resp, err := r.store.Eval(ctx, testScript, []string{r.key}, args).Result()
	if err == redis.Nil {
		return false, err
	} else if err != nil {
		return false, err
	}
	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}
	return exists == 1, nil
}

func (r *redisBitSet) del(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err := r.store.Del(ctx, r.key).Result()
	return err
}

func (r redisBitSet) expire(ctx context.Context, seconds time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err := r.store.Expire(ctx, r.key, seconds).Result()
	return err
}

func (r *redisBitSet) set(ctx context.Context, offsets []uint) error {
	if ctx == nil {
		ctx = context.Background()
	}
	args, err := r.buildOffsetArgs(offsets)
	if err != nil {
		return err
	}
	_, err = r.store.Eval(ctx, setScript, []string{r.key}, args).Result()
	if err == redis.Nil {
		return nil
	} else {
		return err
	}
}