package main

import (
	"context"
	"github.com/ethanvc/study/golangproj/logjson/internal/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"golangbreaker/golangbreaker"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	clearEnv("http_proxy")
	clearEnv("https_proxy")
	runtime.GOMAXPROCS(2)
	http.Handle("/metrics", promhttp.Handler())
	bench := NewBench()
	http.Handle("/", bench)
	addr := ":9100"
	fmt.Println("Starting server on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic("启动 HTTP 服务失败: " + err.Error())
	}
}

var requestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "demo_request_total",
		Help: "Total number of demo requests",
	},
	[]string{"method", "event"},
)

var durationTotal = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "demo_request_duration_seconds",
	},
	[]string{"method"},
)

func init() {
	prometheus.MustRegister(requestTotal, durationTotal)
}

type Bench struct {
	breaker golangbreaker.Breaker
}

func NewBench() *Bench {
	return &Bench{
		breaker: golangbreaker.NewGoSchedBreaker(),
	}
}

func (b *Bench) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.ReadAll(r.Body)
	result := b.Work(r.Context())
	_, _ = w.Write([]byte(result))
}

const ResultOk = "OK"

func (b *Bench) Work(ctx context.Context) string {
	start := time.Now()
	result := b.work(ctx)
	requestTotal.WithLabelValues(b.breaker.Name(), result).Inc()
	durationTotal.WithLabelValues(b.breaker.Name()).Observe(time.Since(start).Seconds())
	return result
}

func (b *Bench) work(ctx context.Context) string {
	if b.breaker.Break() {
		return "FastReject"
	}
	return parseJsonPayload(ctx)
}

func clearEnv(s string) {
	os.Unsetenv(s)
	os.Unsetenv(strings.ToLower(s))
	os.Unsetenv(strings.ToUpper(s))
}

func redisPayload(ctx context.Context) string {
	cmd := sRdb.Get(ctx, "key")
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return "OK"
		}
		return "Error"
	}
	return "OK"
}

var sRdb = redis.NewUniversalClient(&redis.UniversalOptions{
	Addrs: []string{"localhost:6379"}, // Redis 地址
	// 不设置密码
	DB:       0,  // 默认数据库
	PoolSize: 10, // 连接池大小
})

func parseJsonPayload(ctx context.Context) string {
	var val any
	for i := 0; i < 100; i++ {
		err := json.Unmarshal([]byte(sJsonStr), &val)
		if err != nil {
			return "ErrorUnmarshal"
		}
	}
	return ResultOk
}

var sJsonStr = `
{
  "metadata": {
    "version": "1.0.0",
    "generated_at": "2024-01-15T10:30:00Z",
    "description": "大型复杂JSON数据用于压测",
    "data_size": "large",
    "content_type": "application/json"
  },
  "users": [
    {
      "id": 1,
      "personal_info": {
        "name": "Johnathan Smith",
        "email": "johnathan.smith@example.com",
        "age": 32,
        "gender": "male",
        "date_of_birth": "1992-05-15",
        "phone_numbers": [
          {
            "type": "home",
            "number": "+1-555-0123"
          },
          {
            "type": "work",
            "number": "+1-555-0124"
          },
          {
            "type": "mobile",
            "number": "+1-555-0125"
          }
        ],
        "address": {
          "street": "123 Main Street",
          "city": "New York",
          "state": "NY",
          "zip_code": "10001",
          "country": "USA",
          "coordinates": {
            "latitude": 40.7128,
            "longitude": -74.0060
          }
        }
      },
      "account_details": {
        "username": "johnsmith92",
        "password_hash": "5f4dcc3b5aa765d61d8327deb882cf99",
        "account_created": "2020-03-12T14:22:18Z",
        "last_login": "2024-01-15T08:45:23Z",
        "account_status": "active",
        "preferences": {
          "theme": "dark",
          "language": "en-US",
          "notifications": {
            "email": true,
            "sms": false,
            "push": true,
            "frequency": "daily"
          },
          "privacy_settings": {
            "profile_visible": true,
            "search_visible": true,
            "data_sharing": false
          }
        },
        "subscriptions": [
          {
            "service": "premium",
            "start_date": "2023-06-01",
            "end_date": "2024-05-31",
            "auto_renew": true,
            "price": 9.99
          },
          {
            "service": "cloud_storage",
            "start_date": "2023-08-15",
            "end_date": "2024-08-14",
            "auto_renew": true,
            "price": 4.99
          }
        ]
      },
      "social_connections": {
        "friends": [2, 3, 5, 7, 11, 13, 17, 19, 23, 29],
        "followers": [101, 102, 103, 104, 105, 106, 107, 108, 109, 110],
        "following": [201, 202, 203, 204, 205, 206, 207, 208, 209, 210],
        "blocks": []
      },
      "activity_stats": {
        "posts": 142,
        "comments": 567,
        "likes": 2345,
        "shares": 89,
        "views": 123456,
        "engagement_rate": 12.7
      },
      "purchase_history": [
        {
          "order_id": "ORD-2023-001",
          "date": "2023-01-15T10:30:00Z",
          "items": [
            {
              "product_id": "PROD-001",
              "name": "Wireless Headphones",
              "category": "electronics",
              "price": 129.99,
              "quantity": 1,
              "specifications": {
                "color": "black",
                "weight": "250g",
                "battery_life": "20 hours",
                "connectivity": "Bluetooth 5.0"
              }
            },
            {
              "product_id": "PROD-045",
              "name": "Charging Cable",
              "category": "accessories",
              "price": 19.99,
              "quantity": 2,
              "specifications": {
                "length": "2m",
                "type": "USB-C to USB-C",
                "power_delivery": "100W"
              }
            }
          ],
          "total_amount": 169.97,
          "payment_method": "credit_card",
          "shipping_address": {
            "street": "123 Main Street",
            "city": "New York",
            "state": "NY",
            "zip_code": "10001",
            "country": "USA"
          },
          "status": "delivered"
        },
        {
          "order_id": "ORD-2023-078",
          "date": "2023-06-22T14:45:30Z",
          "items": [
            {
              "product_id": "PROD-123",
              "name": "Smart Watch",
              "category": "wearables",
              "price": 249.99,
              "quantity": 1,
              "specifications": {
                "color": "silver",
                "screen_size": "1.5 inch",
                "battery_life": "36 hours",
                "water_resistance": "5 ATM",
                "features": ["heart_rate", "sleep_tracking", "gps", "notifications"]
              }
            }
          ],
          "total_amount": 249.99,
          "payment_method": "paypal",
          "shipping_address": {
            "street": "123 Main Street",
            "city": "New York",
            "state": "NY",
            "zip_code": "10001",
            "country": "USA"
          },
          "status": "delivered"
        }
      ],
      "content_analytics": {
        "top_posts": [
          {
            "post_id": "POST-001",
            "title": "My Vacation in Hawaii",
            "content": "Just returned from an amazing trip to Hawaii...",
            "likes": 245,
            "comments": 32,
            "shares": 15,
            "views": 3456,
            "engagement_rate": 8.4
          },
          {
            "post_id": "POST-045",
            "title": "Tech Review: New Smartphone",
            "content": "Here's my detailed review of the latest smartphone...",
            "likes": 187,
            "comments": 45,
            "shares": 22,
            "views": 5678,
            "engagement_rate": 4.5
          }
        ],
        "content_categories": {
          "travel": 15,
          "technology": 28,
          "food": 12,
          "lifestyle": 25,
          "sports": 8,
          "other": 54
        },
        "audience_demographics": {
          "age_groups": {
            "18-24": 15,
            "25-34": 35,
            "35-44": 25,
            "45-54": 15,
            "55+": 10
          },
          "locations": {
            "USA": 45,
            "UK": 15,
            "Canada": 10,
            "Australia": 8,
            "Germany": 7,
            "Other": 15
          },
          "interests": ["technology", "travel", "photography", "food", "fitness"]
        }
      }
    },
    {
      "id": 2,
      "personal_info": {
        "name": "Emily Johnson",
        "email": "emily.johnson@example.com",
        "age": 28,
        "gender": "female",
        "date_of_birth": "1996-11-23",
        "phone_numbers": [
          {
            "type": "mobile",
            "number": "+1-555-0234"
          }
        ],
        "address": {
          "street": "456 Oak Avenue",
          "city": "Los Angeles",
          "state": "CA",
          "zip_code": "90210",
          "country": "USA",
          "coordinates": {
            "latitude": 34.0522,
            "longitude": -118.2437
          }
        }
      },
      "account_details": {
        "username": "emilyj",
        "password_hash": "5f4dcc3b5aa765d61d8327deb882cf99",
        "account_created": "2021-07-08T09:15:42Z",
        "last_login": "2024-01-15T09:12:37Z",
        "account_status": "active",
        "preferences": {
          "theme": "light",
          "language": "en-US",
          "notifications": {
            "email": true,
            "sms": true,
            "push": true,
            "frequency": "real_time"
          },
          "privacy_settings": {
            "profile_visible": true,
            "search_visible": false,
            "data_sharing": true
          }
        },
        "subscriptions": [
          {
            "service": "basic",
            "start_date": "2023-02-14",
            "end_date": "2024-02-13",
            "auto_renew": false,
            "price": 4.99
          }
        ]
      },
      "social_connections": {
        "friends": [1, 3, 4, 6, 8, 9, 10],
        "followers": [111, 112, 113, 114, 115, 116, 117, 118, 119, 120],
        "following": [211, 212, 213, 214, 215, 216, 217, 218, 219, 220],
        "blocks": [301]
      },
      "activity_stats": {
        "posts": 89,
        "comments": 432,
        "likes": 1876,
        "shares": 67,
        "views": 98765,
        "engagement_rate": 9.2
      },
      "purchase_history": [
        {
          "order_id": "ORD-2023-034",
          "date": "2023-03-10T16:20:15Z",
          "items": [
            {
              "product_id": "PROD-087",
              "name": "Yoga Mat",
              "category": "fitness",
              "price": 39.99,
              "quantity": 1,
              "specifications": {
                "color": "purple",
                "material": "TPE",
                "thickness": "6mm",
                "size": "68x24 inches"
              }
            },
            {
              "product_id": "PROD-088",
              "name": "Yoga Block",
              "category": "fitness",
              "price": 14.99,
              "quantity": 2,
              "specifications": {
                "material": "cork",
                "size": "9x6x4 inches"
              }
            }
          ],
          "total_amount": 69.97,
          "payment_method": "credit_card",
          "shipping_address": {
            "street": "456 Oak Avenue",
            "city": "Los Angeles",
            "state": "CA",
            "zip_code": "90210",
            "country": "USA"
          },
          "status": "delivered"
        }
      ],
      "content_analytics": {
        "top_posts": [
          {
            "post_id": "POST-078",
            "title": "My Yoga Journey",
            "content": "Started practicing yoga 6 months ago and it changed my life...",
            "likes": 167,
            "comments": 28,
            "shares": 12,
            "views": 2345,
            "engagement_rate": 8.9
          },
          {
            "post_id": "POST-092",
            "title": "Healthy Breakfast Ideas",
            "content": "Sharing my favorite healthy breakfast recipes...",
            "likes": 134,
            "comments": 41,
            "shares": 18,
            "views": 3456,
            "engagement_rate": 5.6
          }
        ],
        "content_categories": {
          "health": 22,
          "food": 18,
          "lifestyle": 15,
          "fitness": 28,
          "travel": 7,
          "other": 10
        },
        "audience_demographics": {
          "age_groups": {
            "18-24": 25,
            "25-34": 40,
            "35-44": 20,
            "45-54": 10,
            "55+": 5
          },
          "locations": {
            "USA": 50,
            "UK": 12,
            "Canada": 8,
            "Australia": 7,
            "Germany": 5,
            "Other": 18
          },
          "interests": ["fitness", "health", "food", "wellness", "travel"]
        }
      }
    }
  ],
  "system_metrics": {
    "performance": {
      "response_time": 125,
      "uptime": 99.8,
      "error_rate": 0.2,
      "throughput": 2450,
      "latency": {
        "p50": 45,
        "p90": 120,
        "p95": 180,
        "p99": 350
      }
    },
    "resource_usage": {
      "cpu": 45.2,
      "memory": 67.8,
      "disk": 42.3,
      "network": 12.4
    },
    "database_metrics": {
      "connections": 125,
      "queries_per_second": 345,
      "cache_hit_rate": 87.5,
      "replication_lag": 0
    }
  },
  "analytics": {
    "user_engagement": {
      "daily_active_users": 12456,
      "monthly_active_users": 123456,
      "session_duration": 245,
      "pages_per_session": 4.7,
      "bounce_rate": 32.1
    },
    "content_performance": {
      "top_content": [
        {
          "id": "CONTENT-001",
          "title": "How to Improve Your Productivity",
          "views": 123456,
          "engagement": 12.5,
          "shares": 3456
        },
        {
          "id": "CONTENT-045",
          "title": "10 Tips for Better Sleep",
          "views": 98765,
          "engagement": 10.2,
          "shares": 2345
        },
        {
          "id": "CONTENT-078",
          "title": "Beginner's Guide to Investing",
          "views": 87654,
          "engagement": 9.8,
          "shares": 1987
        }
      ],
      "content_by_category": {
        "technology": 25,
        "health": 18,
        "finance": 15,
        "lifestyle": 22,
        "education": 12,
        "entertainment": 8
      }
    },
    "conversion_metrics": {
      "signup_conversion_rate": 5.2,
      "purchase_conversion_rate": 2.8,
      "retention_rate": 45.7,
      "churn_rate": 3.2,
      "lifetime_value": 125.75
    }
  },
  "additional_data": {
    "feature_flags": {
      "new_ui": true,
      "dark_mode": true,
      "social_sharing": false,
      "advanced_analytics": true,
      "beta_features": false
    },
    "api_endpoints": [
      {
        "name": "user_profile",
        "path": "/api/v1/users/{id}",
        "method": "GET",
        "rate_limit": 1000
      },
      {
        "name": "create_post",
        "path": "/api/v1/posts",
        "method": "POST",
        "rate_limit": 500
      },
      {
        "name": "search",
        "path": "/api/v1/search",
        "method": "GET",
        "rate_limit": 2000
      }
    ],
    "supported_languages": ["en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko"],
    "supported_currencies": ["USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY"],
    "timezones": ["UTC", "EST", "PST", "CET", "JST", "AEST", "IST", "BRT"]
  }
}
`
