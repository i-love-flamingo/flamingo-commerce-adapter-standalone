package domain

import (
	"flamingo.me/flamingo-commerce/v3/category/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCategoryTreeBuilder_BuildTreeWithoutExplicitGivenRoot(t *testing.T) {

	h := &CategoryTreeBuilder{}
	h.AddCategoryData("sub1_sub1", "Sub1 Sub1", "sub1")
	h.AddCategoryData("sub1", "Sub1", "")
	h.AddCategoryData("sub2", "Sub1", "")
	h.AddCategoryData("sub3", "Sub1", "")
	h.AddCategoryData("sub1_sub2_sub1", "Sub1 Sub2 Sub1", "sub1_sub2")
	h.AddCategoryData("sub1_sub2", "Sub1 Sub2", "sub1")

	tree, err := h.BuildTree()
	require.NoError(t, err)
	assert.Equal(t, "", tree.CategoryCode)
	hasSub1 := false
	hasSub2 := false
	hasSub1Sub2 := false
	for _, subt := range tree.SubTrees() {
		if subt.Code() == "sub2" {
			hasSub2 = true
		}
		if subt.Code() == "sub1" {
			hasSub1 = true
			for _, subsubt := range subt.SubTrees() {
				if subsubt.Code() == "sub1_sub2" {
					hasSub1Sub2 = true
				}
			}
		}
	}
	assert.True(t, hasSub1)
	assert.True(t, hasSub2)
	assert.True(t, hasSub1Sub2)

}

func TestCategoryTreeBuilder_BuildTreeWithGivenRoot(t *testing.T) {

	h := &CategoryTreeBuilder{}
	h.AddCategoryData("sub1_sub1", "Sub1 Sub1", "sub1")
	h.AddCategoryData("sub1", "Sub1", "root")
	h.AddCategoryData("sub2", "Sub1", "root")
	h.AddCategoryData("sub3", "Sub1", "root")
	h.AddCategoryData("root", "Root", "root")
	h.AddCategoryData("sub1_sub2_sub1", "Sub1 Sub2 Sub1", "sub1_sub2")
	h.AddCategoryData("sub1_sub2", "Sub1 Sub2", "sub1")

	tree, err := h.BuildTree()
	require.NoError(t, err)
	assert.Equal(t, "root", tree.CategoryCode)
	assert.True(t, tree.HasChilds())
	assert.Equal(t, "Root", tree.Name())
	hasSub1 := false
	hasSub2 := false
	hasSub1Sub2 := false
	for _, subt := range tree.SubTrees() {
		if subt.Code() == "sub2" {
			hasSub2 = true
		}
		if subt.Code() == "sub1" {
			hasSub1 = true
			for _, subsubt := range subt.SubTrees() {
				if subsubt.Code() == "sub1_sub2" {
					hasSub1Sub2 = true
					assert.Equal(t, "/sub1/sub1_sub2", subsubt.Path())
				}
			}
		}
	}
	assert.True(t, hasSub1)
	assert.True(t, hasSub2)
	assert.True(t, hasSub1Sub2)

}

func TestCategoryTreeToCategoryTeaser(t *testing.T) {
	tree := domain.TreeData{
		CategoryCode:          "root",
		CategoryName:          "",
		CategoryPath:          "",
		CategoryDocumentCount: 0,
		SubTreesData: []*domain.TreeData{
			&domain.TreeData{
				CategoryCode:          "sub",
				CategoryName:          "",
				CategoryPath:          "",
				CategoryDocumentCount: 0,
				SubTreesData:          nil,
				IsActive:              false,
			},
		},
		IsActive: false,
	}
	teaser := CategoryTreeToCategoryTeaser("sub", tree)
	require.NotNil(t, teaser)
	assert.Equal(t, "sub", teaser.Code)
	assert.Equal(t, "root", teaser.Parent.Code)
}
