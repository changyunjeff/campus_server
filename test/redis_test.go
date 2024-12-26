package test

import (
	"campus2/pkg/global"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

// TestBasicOperations 测试基本操作
func TestBasicOperations(t *testing.T) {
	rdb := global.GVA_REDIS

	// 测试 String 操作
	t.Run("String Operations", func(t *testing.T) {
		// SET 和 GET
		err := rdb.Set(ctx, "key", "value", time.Hour).Err()
		if err != nil {
			t.Fatalf("Set failed: %v", err)
		}

		val, err := rdb.Get(ctx, "key").Result()
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		t.Logf("get key: %v", val)

		// SETEX (SET with expiration)
		err = rdb.SetEx(ctx, "temp-key", "temp-value", time.Second).Err()
		if err != nil {
			t.Fatalf("SetEx failed: %v", err)
		}

		// 等待过期
		time.Sleep(time.Second * 2)
		_, err = rdb.Get(ctx, "temp-key").Result()
		if err != redis.Nil {
			t.Fatalf("Key should be expired")
		}
		t.Log("Expiration test passed")
	})

	// 测试 Hash 操作
	t.Run("Hash Operations", func(t *testing.T) {
		// HSET 和 HGET
		err := rdb.HSet(ctx, "user:1", "name", "张三", "age", "25").Err()
		if err != nil {
			t.Fatalf("HSet failed: %v", err)
		}

		name, err := rdb.HGet(ctx, "user:1", "name").Result()
		if err != nil {
			t.Fatalf("HGet failed: %v", err)
		}
		t.Logf("user name: %v", name)

		// HGETALL
		fields, err := rdb.HGetAll(ctx, "user:1").Result()
		if err != nil {
			t.Fatalf("HGetAll failed: %v", err)
		}
		t.Logf("user fields: %v", fields)
	})

	// 测试 List 操作
	t.Run("List Operations", func(t *testing.T) {
		// LPUSH 和 RPUSH
		err := rdb.LPush(ctx, "list", "first").Err()
		if err != nil {
			t.Fatalf("LPush failed: %v", err)
		}

		err = rdb.RPush(ctx, "list", "last").Err()
		if err != nil {
			t.Fatalf("RPush failed: %v", err)
		}

		// LRANGE
		values, err := rdb.LRange(ctx, "list", 0, -1).Result()
		if err != nil {
			t.Fatalf("LRange failed: %v", err)
		}
		t.Logf("list values: %v", values)
	})

	// 测试 Set 操作
	t.Run("Set Operations", func(t *testing.T) {
		// SADD
		err := rdb.SAdd(ctx, "set", "member1", "member2", "member3").Err()
		if err != nil {
			t.Fatalf("SAdd failed: %v", err)
		}

		// SMEMBERS
		members, err := rdb.SMembers(ctx, "set").Result()
		if err != nil {
			t.Fatalf("SMembers failed: %v", err)
		}
		t.Logf("set members: %v", members)

		// SISMEMBER
		exists, err := rdb.SIsMember(ctx, "set", "member1").Result()
		if err != nil {
			t.Fatalf("SIsMember failed: %v", err)
		}
		t.Logf("member1 exists: %v", exists)
	})

	// 测试 Sorted Set 操作
	t.Run("Sorted Set Operations", func(t *testing.T) {
		// ZADD
		members := []redis.Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
		}
		err := rdb.ZAdd(ctx, "sorted_set", members...).Err()
		if err != nil {
			t.Fatalf("ZAdd failed: %v", err)
		}

		// ZRANGE
		values, err := rdb.ZRange(ctx, "sorted_set", 0, -1).Result()
		if err != nil {
			t.Fatalf("ZRange failed: %v", err)
		}
		t.Logf("sorted set members: %v", values)
	})
}

// TestTransactions 测试事务
func TestTransactions(t *testing.T) {
	rdb := global.GVA_REDIS

	// 开始事务
	txf := func(tx *redis.Tx) error {
		// 获取计数器的当前值
		n, err := tx.Get(ctx, "counter").Int()
		if err != nil && err != redis.Nil {
			return err
		}

		// 实际的操作
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, "counter", n+1, 0)
			return nil
		})
		return err
	}

	// 使用乐观锁执行事务
	for i := 0; i < 3; i++ {
		err := rdb.Watch(ctx, txf, "counter")
		if err != nil {
			t.Fatalf("Transaction failed: %v", err)
		}
	}

	// 验证结果
	val, err := rdb.Get(ctx, "counter").Int()
	if err != nil {
		t.Fatalf("Get counter failed: %v", err)
	}
	t.Logf("Final counter value: %d", val)
}

// TestPipeline 测试管道
func TestPipeline(t *testing.T) {
	rdb := global.GVA_REDIS

	pipe := rdb.Pipeline()

	// 将多个命令放入管道
	incr := pipe.Incr(ctx, "pipeline_counter")
	pipe.Expire(ctx, "pipeline_counter", time.Hour)
	pipe.Set(ctx, "pipeline_key", "pipeline_value", time.Hour)
	get := pipe.Get(ctx, "pipeline_key")

	// 执行管道中的所有命令
	_, err := pipe.Exec(ctx)
	if err != nil {
		t.Fatalf("Pipeline execution failed: %v", err)
	}

	// 获取命令的结果
	counter, err := incr.Result()
	if err != nil {
		t.Fatalf("Failed to get incr result: %v", err)
	}
	t.Logf("Pipeline counter: %d", counter)

	value, err := get.Result()
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	t.Logf("Pipeline value: %s", value)
}

// TestPubSub 测试发布订阅
func TestPubSub(t *testing.T) {
	rdb := global.GVA_REDIS

	// 订阅频道
	pubsub := rdb.Subscribe(ctx, "mychannel")
	defer pubsub.Close()

	// 在goroutine中处理接收到的消息
	done := make(chan bool)
	go func() {
		defer close(done)

		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			t.Errorf("Failed to receive message: %v", err)
			return
		}
		t.Logf("Received message from %s: %s", msg.Channel, msg.Payload)
		done <- true
	}()

	// 发布消息
	err := rdb.Publish(ctx, "mychannel", "Hello Redis PubSub").Err()
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// 等待消息接收
	select {
	case <-done:
		t.Log("PubSub test completed")
	case <-time.After(time.Second * 5):
		t.Fatal("PubSub test timed out")
	}
}

// TestCustomDuration 测试自定义过期时间
func TestCustomDuration(t *testing.T) {
	rdb := global.GVA_REDIS

	// 使用配置的默认过期时间
	err := rdb.Set(ctx, "custom_duration_key", "value", global.GVA_CONFIG.Redis.GetDuration()).Err()
	if err != nil {
		t.Fatalf("Set with custom duration failed: %v", err)
	}

	// 获取过期时间
	ttl, err := rdb.TTL(ctx, "custom_duration_key").Result()
	if err != nil {
		t.Fatalf("TTL check failed: %v", err)
	}
	t.Logf("Key TTL: %v", ttl)

	// 验证值
	val, err := rdb.Get(ctx, "custom_duration_key").Result()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	t.Logf("Value: %v", val)
}

// TestCleanup 清理测试数据
func TestCleanup(t *testing.T) {
	rdb := global.GVA_REDIS

	// 列出所有测试使用的键
	keys := []string{
		"key",
		"temp-key",
		"user:1",
		"list",
		"set",
		"sorted_set",
		"counter",
		"pipeline_counter",
		"pipeline_key",
		"custom_duration_key",
	}

	// 删除所有测试键
	for _, key := range keys {
		err := rdb.Del(ctx, key).Err()
		if err != nil {
			t.Logf("Failed to delete key %s: %v", key, err)
		}
	}

	t.Log("Cleanup completed")
}

// 用于打印Redis信息的辅助函数
func printRedisInfo(t *testing.T) {
	rdb := global.GVA_REDIS
	info, err := rdb.Info(ctx).Result()
	if err != nil {
		t.Logf("Failed to get Redis info: %v", err)
		return
	}
	fmt.Printf("Redis Info:\n%s\n", info)
}

// TestConnection 测试连接
func TestConnection(t *testing.T) {
	rdb := global.GVA_REDIS

	// 测试 PING
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis连接测试失败: %v", err)
	}
	t.Logf("Redis连接测试成功: %v", pong)

	// 获取一些服务器信息
	info, err := rdb.Info(ctx, "server").Result()
	if err != nil {
		t.Fatalf("获取Redis信息失败: %v", err)
	}
	t.Logf("Redis服务器信息: \n%s", info)
}
