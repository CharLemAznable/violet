package violet_test

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/CharLemAznable/violet"
	"os"
	"testing"
)

func TestConfigHelpInfo(t *testing.T) {
	config := &violet.Config{
		Endpoint: []violet.EndpointConfig{
			{
				Name:                "指定代理端点名称",
				Location:            "代理匹配路径",
				StripLocationPrefix: "访问代理目标时是否去除匹配路径前缀, true/false",
				TargetURL:           "代理目标地址",
				DumpTarget:          "是否记录访问代理目标的请求响应事件, true/false",
				DumpSource:          "是否记录访问代理来源的请求响应事件, true/false",
			},
		},
		Defaults: violet.Defaults{
			Resilience: violet.ResilienceConfig{
				Bulkhead: violet.BulkheadConfig{
					Disabled:           "舱壁隔离开关, true/false",
					MaxConcurrentCalls: "最大并发量, 默认值: 25",
					MaxWaitDuration:    "当并发超量时的最大等待时长, 默认值: 0",
					WhenFullResponse:   "当并发超量时的fallback响应, 要求为完整的HTTP报文格式",
					Order:              "调用的包装顺序, 默认值: '100'",
				},
				TimeLimiter: violet.TimeLimiterConfig{
					Disabled:            "调用时长限制开关, true/false",
					TimeoutDuration:     "调用时限, 默认值: 60s",
					WhenTimeoutResponse: "当调用超时时的fallback响应, 要求为完整的HTTP报文格式",
					Order:               "调用的包装顺序, 默认值: '200'",
				},
				RateLimiter: violet.RateLimiterConfig{
					Disabled:             "调用频率限制开关, true/false",
					TimeoutDuration:      "等待允许调用的最大等待时长, 默认值: 5s",
					LimitRefreshPeriod:   "并发数量的刷新时间, 默认值: 500ns",
					LimitForPeriod:       "刷新时间内允许的并发数量, 默认值: 50",
					WhenOverRateResponse: "当调用超速时的fallback响应, 要求为完整的HTTP报文格式",
					Order:                "调用的包装顺序, 默认值: '300'",
				},
				CircuitBreaker: violet.CircuitBreakerConfig{
					Disabled:                  "断路器开关, true/false",
					SlidingWindowType:         "滑动窗口类型, TIME_BASED/COUNT_BASED, 默认值: COUNT_BASED",
					SlidingWindowSize:         "滑动窗口大小, 默认值: 100",
					MinimumNumberOfCalls:      "断路器判断启动的最小调用次数, 默认值: 100",
					FailureRateThreshold:      "断路开启的失败调用率阈值百分比, 0~100, 默认值: 50",
					SlowCallRateThreshold:     "断路开启的慢调用率阈值百分比, 0~100, 默认值: 100",
					SlowCallDurationThreshold: "慢调用判断的时长阈值, 默认值: 60s",
					ResponseFailedPredicate:   "失败调用的断言器名称, 使用RegisterRspFailedPredicate方法注册",
					ResponseFailedPredicateContext: map[string]string{
						"断言器上下文参数": "可自定义断言器的判断逻辑",
					},
					AutomaticTransitionFromOpenToHalfOpen: "是否自动从断路开启转换为断路半开, true/false, 默认值: false",
					WaitIntervalInOpenState:               "自动从断路开启转换为断路半开的等待时长, 默认值: 60s",
					PermittedNumberOfCallsInHalfOpenState: "断路半开时允许通过的调用次数, 默认值: 10",
					MaxWaitDurationInHalfOpenState:        "断路半开时的最大等待时长, 默认值: 0",
					WhenOverLoadResponse:                  "当断路开启时的fallback响应, 要求为完整的HTTP报文格式",
					Order:                                 "调用的包装顺序, 默认值: '400'",
				},
				Retry: violet.RetryConfig{
					Disabled:                "重试器开关, true/false",
					MaxAttempts:             "最大重试次数, 默认值: 3",
					FailAfterMaxAttempts:    "是否在最后一次重试失败后返回错误, true/false, 默认值: false",
					ResponseFailedPredicate: "失败调用的断言器名称, 使用RegisterRspFailedPredicate方法注册",
					ResponseFailedPredicateContext: map[string]string{
						"断言器上下文参数": "可自定义断言器的判断逻辑",
					},
					WaitInterval:           "重试的等待时长, 默认值: 500ms",
					WhenMaxRetriesResponse: "当最后一次重试失败后返回错误时的fallback响应, 要求为完整的HTTP报文格式",
					Order:                  "调用的包装顺序, 默认值: '500'",
				},
				Cache: violet.CacheConfig{
					Enabled:                "缓存开关, true/false",
					Capacity:               "缓存容量, 默认值: 10000",
					ItemTTL:                "缓存有效时间, 默认值: 5m",
					ResponseCachePredicate: "是否缓存响应的断言器名称, 使用RegisterRspCachePredicate方法注册",
					ResponseCachePredicateContext: map[string]string{
						"断言器上下文参数": "可自定义断言器的判断逻辑",
					},
					Order: "调用的包装顺序, 默认值: '600'",
				},
				Fallback: violet.FallbackConfig{
					Enabled:          "故障恢复开关, true/false",
					FallbackResponse: "故障恢复时的fallback响应, 要求为完整的HTTP报文格式",
					FallbackFunction: "故障恢复的函数名称, 使用RegisterFallbackFunction方法注册",
					FallbackFunctionContext: map[string]string{
						"函数上下文参数": "可自定义函数的处理逻辑",
					},
					ResponseFailedPredicate: "是否发生故障的断言器名称, 使用RegisterRspCachePredicate方法注册",
					ResponseFailedPredicateContext: map[string]string{
						"断言器上下文参数": "可自定义断言器的判断逻辑",
					},
					Order: "调用的包装顺序, 默认值: '700'",
				},
			},
		},
	}
	formatConfig := violet.FormatConfig(config)
	buffer := new(bytes.Buffer)
	encoder := toml.NewEncoder(buffer)
	_ = encoder.Encode(formatConfig)
	content := new(bytes.Buffer)
	scanner := bufio.NewScanner(buffer)
	for scanner.Scan() {
		line := scanner.Text()
		content.WriteString("# " + line + "\n")
	}
	file, err := os.Create("config_help.toml")
	if err != nil {
		fmt.Println("创建文件时发生错误:", err)
		return
	}
	defer func() { _ = file.Close() }()
	_, _ = file.WriteString(content.String())
}
