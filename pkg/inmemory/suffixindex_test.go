package inmemory_test

import (
	"context"
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"

	"github.com/dhlab-tech/go-mongo-platform/pkg/inmemory"
)

func TestS_Put_1(t *testing.T) {
	a := inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool())
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	res := a.Search("Було")
	assert.Equal(t, []int{1}, res)
}

func TestS_Put_2(t *testing.T) {
	a := inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool())
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	a.Put("Булочка с корицей", 3)
	res := a.Search("Було")
	assert.Equal(t, []int{1, 3}, res)
}

func TestS_Put_3(t *testing.T) {
	a := inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool())
	a.Put("Булочка с вишней", 1)
	a.Put("БАК ФАСОВКА Булгур", 2)
	a.Put("Булочка с корицей", 3)
	res := a.Search("фас")
	assert.Equal(t, []int{2}, res)
	res = a.Search("кор")
	assert.Equal(t, []int{3}, res)
}

func TestM_S(t *testing.T) {
	c := inmemory.NewCache[*inmemory.Image](make(map[string]*inmemory.Image))
	m := inmemory.NewM(
		c,
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
	)
	expected := []string{}
	for _, v := range []string{"Выпечка", "Выпечка сладкая", "Выпечка сытная"} {
		v := v
		im := inmemory.Image{Name: &v}
		c.Add(context.Background(), &im)
		m.Add(im.ID(), v)
		expected = append(expected, im.ID())
	}
	res := m.S(context.Background(), "выпе")
	assert.Equal(t, expected, res)
	res = m.S(context.Background(), "дкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = m.S(context.Background(), "сладкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = m.S(context.Background(), "ечка")
	assert.Equal(t, expected, res)
}

func TestM_Rebuild(t *testing.T) {
	c := inmemory.NewCache[*inmemory.Image](make(map[string]*inmemory.Image))
	m := inmemory.NewM(
		c,
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
	)
	m.Start()
	expected := []string{}
	for _, v := range []string{"Выпечка", "Выпечка сладкая", "Выпечка сытная"} {
		v := v
		im := inmemory.Image{Name: &v}
		c.Add(context.Background(), &im)
		m.Rebuild(im.ID(), v)
		expected = append(expected, im.ID())
	}
	m.Commit()
	res := m.S(context.Background(), "выпе")
	assert.Equal(t, expected, res)
	res = m.S(context.Background(), "дкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = m.S(context.Background(), "сладкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = m.S(context.Background(), "ечка")
	assert.Equal(t, expected, res)
}

func TestSuffix_Rebuild(t *testing.T) {
	c := inmemory.NewCache[*inmemory.Image](make(map[string]*inmemory.Image))
	m := inmemory.NewM(
		c,
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
		inmemory.NewS(inmemory.NewIntersect(), btree.New(100), inmemory.NewPool()),
	)
	s := inmemory.NewSuffix(m, c, []string{"Name"}, nil)
	expected := []string{}
	for _, v := range []string{"Выпечка", "Выпечка сладкая", "Выпечка сытная"} {
		v := v
		im := inmemory.Image{Name: &v}
		c.Add(context.Background(), &im)
		s.Add(context.Background(), &im)
		expected = append(expected, im.ID())
	}
	res := s.Search(context.Background(), "выпе")
	assert.Equal(t, expected, res)
	res = s.Search(context.Background(), "дкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = s.Search(context.Background(), "сладкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = s.Search(context.Background(), "ечка")
	assert.Equal(t, expected, res)
	s.Rebuild(context.Background())
	res = s.Search(context.Background(), "выпе")
	assert.Equal(t, expected, res)
	res = s.Search(context.Background(), "дкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = s.Search(context.Background(), "сладкая")
	assert.Equal(t, []string{expected[1]}, res)
	res = s.Search(context.Background(), "ечка")
	assert.Equal(t, expected, res)
}
