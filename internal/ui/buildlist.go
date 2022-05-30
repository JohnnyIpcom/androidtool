package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/generic"
	"golang.org/x/sync/errgroup"
)

type BuildType int

const (
	BuildTypeAPK BuildType = iota
	BuildTypeAAB
)

func (b BuildType) String() string {
	switch b {
	case BuildTypeAPK:
		return "APK"
	case BuildTypeAAB:
		return "AAB"
	}
	return ""
}

func (b BuildType) Icon() fyne.Resource {
	switch b {
	case BuildTypeAPK:
		return assets.APKIcon
	case BuildTypeAAB:
		return assets.AABIcon
	}
	return nil
}

// Build is a struct that contains the information about a build.
type Build struct {
	Type         BuildType
	Path         string
	APKsPath     string
	UnpackedPath string

	abi      *widget.Button
	sizes    *widget.Button
	manifest *widget.Button
}

type BuildList struct {
	widget.List

	aabClient *aabclient.Client
	aapt      *aapt.AAPT
	items     *generic.Slice[*Build]
	parent    fyne.Window
}

func (b *BuildList) Length() int {
	return b.items.Len()
}

func (b *BuildList) CreateItem() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		container.NewHBox(
			widget.NewIcon(assets.BuildsTabIcon),
			widget.NewLabel("<BUILD>"),
		),
		container.NewHBox(
			widget.NewButtonWithIcon("", assets.ABIIcon, nil),
			widget.NewButtonWithIcon("", assets.SizesIcon, nil),
			widget.NewButtonWithIcon("", assets.ManifestIcon, nil),
		),
	)
}

func (b *BuildList) UpdateItem(id int, item fyne.CanvasObject) {
	c := item.(*fyne.Container)

	buildItem := b.items.Load(id)
	c.Objects[0].(*fyne.Container).Objects[0].(*widget.Icon).SetResource(buildItem.Type.Icon())
	c.Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(buildItem.Path)
	buildItem.abi = c.Objects[1].(*fyne.Container).Objects[0].(*widget.Button)
	buildItem.abi.OnTapped = func() {
		go func() {
			switch buildItem.Type {
			case BuildTypeAPK:
				APKABIInfo(b.aapt, buildItem.Path, b.parent)

			case BuildTypeAAB:
				AABABIInfo(b.aabClient, b.aapt, buildItem.UnpackedPath, b.parent)

			default:
				ShowError(fmt.Errorf("unknown build type: %s", buildItem.Type), nil, b.parent)
			}
		}()
	}

	buildItem.sizes = c.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)
	buildItem.sizes.OnTapped = func() {
		go func() {
			switch buildItem.Type {
			case BuildTypeAPK:
				APKSizes(buildItem.Path, b.parent)

			case BuildTypeAAB:
				AABSizes(b.aabClient, buildItem.APKsPath, b.parent)

			default:
				ShowError(fmt.Errorf("unknown build type: %s", buildItem.Type), nil, b.parent)
			}
		}()
	}

	buildItem.manifest = c.Objects[1].(*fyne.Container).Objects[2].(*widget.Button)
	buildItem.manifest.OnTapped = func() {
		go func() {
			switch buildItem.Type {
			case BuildTypeAPK:
				APKManifest(b.aapt, buildItem.Path, b.parent)

			case BuildTypeAAB:
				AABManifest(b.aabClient, buildItem.Path, b.parent)

			default:
				ShowError(fmt.Errorf("unknown build type: %s", buildItem.Type), nil, b.parent)
			}
		}()
	}
}

func (b *BuildList) OnSelected(id int) {
	b.Unselect(id)
}

func (b *BuildList) LoadAPK(path string) {
	b.items.Store(
		&Build{
			Type: BuildTypeAPK,
			Path: path,
		},
	)

	b.Refresh()
}

func (b *BuildList) LoadAAB(path string, useCachedData bool) {
	var aabInfo *AABInfo

	g := errgroup.Group{}
	g.Go(func() error {
		info, err := LoadAAB(b.aabClient, b.aapt, path, useCachedData, b.parent)
		if err != nil {
			return err
		}

		aabInfo = info
		return nil
	})

	if err := g.Wait(); err != nil {
		ShowError(err, nil, b.parent)
		return
	}

	b.items.Store(
		&Build{
			Type:         BuildTypeAAB,
			Path:         path,
			APKsPath:     aabInfo.APKsPath,
			UnpackedPath: aabInfo.UnpackedPath,
		},
	)

	b.Refresh()
}

func NewBuildList(aabClient *aabclient.Client, aapt *aapt.AAPT, parent fyne.Window) *BuildList {
	d := &BuildList{
		aabClient: aabClient,
		aapt:      aapt,
		items:     generic.NewSlice[*Build](),
		parent:    parent,
	}

	d.List.Length = d.Length
	d.List.CreateItem = d.CreateItem
	d.List.UpdateItem = d.UpdateItem
	d.List.OnSelected = d.OnSelected

	return d
}
