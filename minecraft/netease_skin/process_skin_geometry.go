/*
PhoenixBuilder specific packages.
Author: Liliya233, Happy2018new
*/
package NetEaseSkin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	_ "embed"
)

type SkinCube struct {
	Inflate *json.Number  `json:"inflate,omitempty"`
	Mirror  *bool         `json:"mirror,omitempty"`
	Origin  []json.Number `json:"origin"`
	Size    []json.Number `json:"size"`
	Uv      []json.Number `json:"uv"`
}

type SkinGeometryBone struct {
	Cubes         *[]SkinCube   `json:"cubes,omitempty"`
	Name          string        `json:"name"`
	Parent        string        `json:"parent,omitempty"`
	Pivot         []json.Number `json:"pivot"`
	RenderGroupID int           `json:"render_group_id,omitempty"`
	Rotation      []json.Number `json:"rotation,omitempty"`
}

type SkinGeometry struct {
	Bones               []*SkinGeometryBone `json:"bones"`
	TextureHeight       int                 `json:"textureheight"`
	TextureWidth        int                 `json:"texturewidth"`
	VisibleBoundsHeight json.Number         `json:"visible_bounds_height,omitempty"`
	VisibleBoundsOffset []json.Number       `json:"visible_bounds_offset,omitempty"`
	VisibleBoundsWidth  json.Number         `json:"visible_bounds_width,omitempty"`
}

func ProcessGeometry(skin *Skin, rawData []byte) (err error) {
	/* Layer 1 */
	geometryMap := map[string]json.RawMessage{}
	if err = json.Unmarshal(rawData, &geometryMap); err != nil {
		return fmt.Errorf("ProcessGeometry: %v", err)
	}
	// setup resource patch and get geometry data
	var skinGeometry json.RawMessage
	var geometryName string
	for k, v := range geometryMap {
		if strings.HasPrefix(k, "geometry.") {
			geometryName = k
			skinGeometry = v
			break
		}
	}
	if geometryName == "" {
		return fmt.Errorf("ProcessGeometry: lack of geometry data")
	}
	skin.SkinResourcePatch = bytes.ReplaceAll(
		skin.SkinResourcePatch,
		[]byte("geometry.humanoid.custom"),
		[]byte(geometryName),
	)
	/* Layer 2 */
	geometry := &SkinGeometry{}
	if err = json.Unmarshal(skinGeometry, geometry); err != nil {
		return fmt.Errorf("ProcessGeometry: %v", err)
	}
	// handle bones
	hasRoot := false
	renderGroupNames := []string{"leftArm", "rightArm"}
	for _, bone := range geometry.Bones {
		// setup parent
		switch bone.Name {
		case "waist", "leftLeg", "rightLeg":
			bone.Parent = "root"
		case "head":
			bone.Parent = "body"
		case "leftArm", "rightArm":
			bone.Parent = "body"
			bone.RenderGroupID = 1
		case "body":
			bone.Parent = "waist"
		case "root":
			hasRoot = true
		}
		// setup render group
		if slices.Contains(renderGroupNames, bone.Parent) {
			bone.RenderGroupID = 1
			renderGroupNames = append(renderGroupNames, bone.Name)
		}
	}
	if !hasRoot {
		geometry.Bones = append(geometry.Bones, &SkinGeometryBone{
			Name: "root",
			Pivot: []json.Number{
				json.Number("0.0"),
				json.Number("0.0"),
				json.Number("0.0"),
			},
		})
	}
	// return
	skin.SkinGeometry, _ = json.Marshal(map[string]any{
		"format_version": "1.8.0",
		geometryName:     geometry,
	})
	return
}
