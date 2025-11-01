package prefixmatch

import (
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	if tree == nil {
		t.Fatal("NewTree should return a non-nil tree")
	}
	if tree.children == nil {
		t.Error("NewTree should initialize children map")
	}
	if tree.emptyNode != nil {
		t.Error("NewTree should have nil emptyNode initially")
	}
}

func TestInsert_EmptyPattern(t *testing.T) {
	tree := NewTree()

	// 插入空模式
	err := tree.Insert("", 100)
	if err != nil {
		t.Errorf("Insert empty pattern should succeed, got error: %v", err)
	}

	// 再次插入空模式应该失败
	err = tree.Insert("", 200)
	if err == nil {
		t.Error("Insert duplicate empty pattern should fail")
	}
}

func TestInsert_SinglePattern(t *testing.T) {
	tree := NewTree()

	err := tree.Insert("test", 1)
	if err != nil {
		t.Errorf("Insert should succeed, got error: %v", err)
	}

	// 验证可以搜索到
	node := tree.Search("test")
	if node == nil {
		t.Error("Should find inserted pattern")
	}
	if node.value != 1 {
		t.Errorf("Expected value 1, got %d", node.value)
	}
}

func TestInsert_MultiplePatterns(t *testing.T) {
	tree := NewTree()

	patterns := map[string]int{
		"test":    1,
		"team":    2,
		"toast":   3,
		"testing": 4,
	}

	for pattern, value := range patterns {
		err := tree.Insert(pattern, value)
		if err != nil {
			t.Errorf("Insert %s failed: %v", pattern, err)
		}
	}

	// 验证所有模式都能找到
	for pattern, expectedValue := range patterns {
		node := tree.Search(pattern)
		if node == nil {
			t.Errorf("Should find pattern: %s", pattern)
			continue
		}
		if node.value != expectedValue {
			t.Errorf("Pattern %s: expected value %d, got %d", pattern, expectedValue, node.value)
		}
	}
}

func TestInsert_DuplicatePattern(t *testing.T) {
	tree := NewTree()

	err := tree.Insert("test", 1)
	if err != nil {
		t.Errorf("First insert should succeed, got error: %v", err)
	}

	// 重复插入相同的模式应该失败
	err = tree.Insert("test", 2)
	if err == nil {
		t.Error("Insert duplicate pattern should fail")
	}
}

func TestInsert_WithCommonPrefix(t *testing.T) {
	tree := NewTree()

	// 插入有公共前缀的模式
	err := tree.Insert("test", 1)
	if err != nil {
		t.Errorf("Insert 'test' failed: %v", err)
	}

	err = tree.Insert("testing", 2)
	if err != nil {
		t.Errorf("Insert 'testing' failed: %v", err)
	}

	err = tree.Insert("tester", 3)
	if err != nil {
		t.Errorf("Insert 'tester' failed: %v", err)
	}

	// 验证所有模式都能正确找到
	tests := []struct {
		pattern string
		value   int
	}{
		{"test", 1},
		{"testing", 2},
		{"tester", 3},
	}

	for _, tt := range tests {
		node := tree.Search(tt.pattern)
		if node == nil {
			t.Errorf("Should find pattern: %s", tt.pattern)
			continue
		}
		if node.value != tt.value {
			t.Errorf("Pattern %s: expected value %d, got %d", tt.pattern, tt.value, node.value)
		}
	}
}

func TestInsert_PrefixIsKey(t *testing.T) {
	tree := NewTree()

	// 先插入长的
	err := tree.Insert("testing", 1)
	if err != nil {
		t.Errorf("Insert 'testing' failed: %v", err)
	}

	// 再插入短的（前缀）
	err = tree.Insert("test", 2)
	if err != nil {
		t.Errorf("Insert 'test' failed: %v", err)
	}

	// 两个都应该能找到
	node := tree.Search("testing")
	if node == nil || node.value != 1 {
		t.Error("Should find 'testing' with value 1")
	}

	node = tree.Search("test")
	if node == nil || node.value != 2 {
		t.Error("Should find 'test' with value 2")
	}
}

func TestInsert_NodeSplitting(t *testing.T) {
	tree := NewTree()

	// 插入导致节点分裂的模式
	err := tree.Insert("romane", 1)
	if err != nil {
		t.Errorf("Insert 'romane' failed: %v", err)
	}

	err = tree.Insert("romanus", 2)
	if err != nil {
		t.Errorf("Insert 'romanus' failed: %v", err)
	}

	err = tree.Insert("rubens", 3)
	if err != nil {
		t.Errorf("Insert 'rubens' failed: %v", err)
	}

	// 验证所有模式
	tests := []struct {
		pattern string
		value   int
	}{
		{"romane", 1},
		{"romanus", 2},
		{"rubens", 3},
	}

	for _, tt := range tests {
		node := tree.Search(tt.pattern)
		if node == nil {
			t.Errorf("Should find pattern: %s", tt.pattern)
			continue
		}
		if node.value != tt.value {
			t.Errorf("Pattern %s: expected value %d, got %d", tt.pattern, tt.value, node.value)
		}
	}
}

func TestSearch_NotFound(t *testing.T) {
	tree := NewTree()

	tree.Insert("test", 1)
	tree.Insert("team", 2)

	// 搜索不存在的模式
	tests := []string{
		"notexist",
		"te",
		"testing",
		"teams",
		"toast",
	}

	for _, pattern := range tests {
		node := tree.Search(pattern)
		if node != nil {
			t.Errorf("Should not find pattern: %s", pattern)
		}
	}
}

func TestSearch_EmptyPattern(t *testing.T) {
	tree := NewTree()

	// 没有插入空模式时搜索
	node := tree.Search("")
	if node != nil {
		t.Error("Should not find empty pattern when not inserted")
	}

	// 插入空模式后搜索
	tree.Insert("", 100)
	node = tree.Search("")
	if node == nil {
		t.Error("Should find empty pattern after insert")
	}
	if node.value != 100 {
		t.Errorf("Expected value 100, got %d", node.value)
	}
}

func TestSearch_SingleCharacter(t *testing.T) {
	tree := NewTree()

	tree.Insert("a", 1)
	tree.Insert("b", 2)
	tree.Insert("c", 3)

	node := tree.Search("a")
	if node == nil || node.value != 1 {
		t.Error("Should find 'a' with value 1")
	}

	node = tree.Search("b")
	if node == nil || node.value != 2 {
		t.Error("Should find 'b' with value 2")
	}

	node = tree.Search("c")
	if node == nil || node.value != 3 {
		t.Error("Should find 'c' with value 3")
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected string
	}{
		{"test", "testing", "test"},
		{"hello", "world", ""},
		{"abc", "abc", "abc"},
		{"", "test", ""},
		{"test", "", ""},
		{"prefix", "preorder", "pre"},
		{"a", "b", ""},
	}

	for _, tt := range tests {
		result := longestCommonPrefix(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("longestCommonPrefix(%q, %q) = %q, want %q", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestNode_Reset(t *testing.T) {
	node := &Node{
		part:       "test",
		pattern:    "testing",
		children:   map[byte]*Node{'a': {}},
		value:      100,
		valueValid: true,
	}

	node.reset()

	if node.part != "" {
		t.Error("part should be empty after reset")
	}
	if node.pattern != "" {
		t.Error("pattern should be empty after reset")
	}
	if node.value != 0 {
		t.Error("value should be 0 after reset")
	}
	if node.valueValid {
		t.Error("valueValid should be false after reset")
	}
	if node.children == nil {
		t.Error("children should not be nil after reset")
	}
}

func TestNode_CandidateChild(t *testing.T) {
	node := &Node{
		children: map[byte]*Node{
			'a': {part: "apple"},
			'b': {part: "banana"},
		},
	}

	child := node.candidateChild("apple")
	if child == nil || child.part != "apple" {
		t.Error("Should find child starting with 'a'")
	}

	child = node.candidateChild("banana")
	if child == nil || child.part != "banana" {
		t.Error("Should find child starting with 'b'")
	}

	child = node.candidateChild("cherry")
	if child != nil {
		t.Error("Should not find child starting with 'c'")
	}
}

func TestNewNode(t *testing.T) {
	node := newNode("testing", "ing", 42)

	if node.pattern != "testing" {
		t.Errorf("Expected pattern 'testing', got %s", node.pattern)
	}
	if node.part != "ing" {
		t.Errorf("Expected part 'ing', got %s", node.part)
	}
	if node.value != 42 {
		t.Errorf("Expected value 42, got %d", node.value)
	}
	if !node.valueValid {
		t.Error("valueValid should be true")
	}
}

func TestComplexScenario(t *testing.T) {
	tree := NewTree()

	// 插入一系列复杂的模式
	patterns := []struct {
		pattern string
		value   int
	}{
		{"romane", 1},
		{"romanus", 2},
		{"romulus", 3},
		{"rubens", 4},
		{"ruber", 5},
		{"rubicon", 6},
		{"rubicundus", 7},
	}

	for _, p := range patterns {
		err := tree.Insert(p.pattern, p.value)
		if err != nil {
			t.Errorf("Insert %s failed: %v", p.pattern, err)
		}
	}

	// 验证所有模式都能正确找到
	for _, p := range patterns {
		node := tree.Search(p.pattern)
		if node == nil {
			t.Errorf("Should find pattern: %s", p.pattern)
			continue
		}
		if node.value != p.value {
			t.Errorf("Pattern %s: expected value %d, got %d", p.pattern, p.value, node.value)
		}
	}

	// 验证不存在的模式
	notExist := []string{"rom", "rub", "rubic", "roma", "ruben"}
	for _, pattern := range notExist {
		node := tree.Search(pattern)
		if node != nil {
			t.Errorf("Should not find pattern: %s", pattern)
		}
	}
}

func TestTreeSearch_Internal(t *testing.T) {
	tree := NewTree()

	tree.Insert("testing", 1)
	tree.Insert("test", 2)
	tree.Insert("team", 3)

	// 测试内部 search 方法
	node, rest := tree.search("testing")
	if node == nil {
		t.Error("Should find node for 'testing'")
	}
	if rest != "" {
		t.Errorf("Expected empty rest, got %s", rest)
	}

	node, rest = tree.search("test")
	if node == nil {
		t.Error("Should find node for 'test'")
	}
	if rest != "" {
		t.Errorf("Expected empty rest, got %s", rest)
	}

	// 搜索不存在但有部分匹配的
	node, rest = tree.search("testimony")
	if node == nil {
		t.Error("Should find partial match node")
	}
	if rest == "" {
		t.Error("Should have remaining pattern")
	}
}
