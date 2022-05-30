package ui

import (
	"fmt"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/c2h5oh/datasize"
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
	Type     BuildType
	Path     string
	APKsPath string
	ABIList  []string
	MinSize  datasize.ByteSize
	MaxSize  datasize.ByteSize

	abi      *widget.Button
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
		go b.onABIButtonTapped(buildItem)
	}

	buildItem.manifest = c.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)
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

func (b *BuildList) onABIButtonTapped(buildItem *Build) {
	if len(buildItem.ABIList) == 0 {
		ShowInformation("ABI Info", "No ABIs found", b.parent)
		return
	}

	data := binding.BindStringList(&buildItem.ABIList)
	list := widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewLabel("<ABI>")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(310, 200))

	minSizeLabel := widget.NewLabel(fmt.Sprintf("%s (%d)", buildItem.MinSize.HumanReadable(), buildItem.MinSize.Bytes()))
	maxSizeLabel := widget.NewLabel(fmt.Sprintf("%s (%d)", buildItem.MaxSize.HumanReadable(), buildItem.MaxSize.Bytes()))

	dialog.ShowCustom(
		"ABI Info",
		"Close",
		container.NewVBox(
			widget.NewCard(
				"",
				"ABIs:",
				container.NewMax(
					rect,
					list,
				),
			),
			widget.NewCard(
				"",
				"Sizes:",
				container.NewGridWithColumns(
					2,
					widget.NewLabel("Min:"),
					minSizeLabel,
					widget.NewLabel("Max:"),
					maxSizeLabel,
				),
			),
		),
		b.parent,
	)
}

func (b *BuildList) LoadAPK(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		ShowError(err, nil, b.parent)
		return
	}

	abis, err := b.aapt.GetNativeCodeABIs(path)
	if err != nil {
		ShowError(err, nil, b.parent)
		return
	}

	b.items.Store(
		&Build{
			Type:    BuildTypeAPK,
			Path:    path,
			ABIList: abis,
			MinSize: datasize.ByteSize(stat.Size()),
			MaxSize: datasize.ByteSize(stat.Size()),
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
			Type:    BuildTypeAAB,
			Path:    path,
			ABIList: aabInfo.ABIList,
			MinSize: datasize.ByteSize(aabInfo.MinSize),
			MaxSize: datasize.ByteSize(aabInfo.MaxSize),
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
