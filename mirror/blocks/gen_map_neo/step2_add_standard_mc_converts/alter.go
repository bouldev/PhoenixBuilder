package step2_add_standard_mc_converts

import "strings"

func AlterInOutSnbtBlock(inBlockName, inBlockState, outBlockName, outBlockState string) (string, string, string, string) {
	if outBlockName == "cherry_sign" {
		// fix
		outBlockName = "standing_sign"
		outBlockState = strings.ReplaceAll(outBlockState, "rotation", "ground_sign_direction")
	} else if outBlockName == "cherry_wall_sign" {
		outBlockName = "wall_sign"
		// outBlockState = strings.ReplaceAll(outBlockState, "facing", "facing_direction")
		// outBlockState = strings.ReplaceAll(outBlockState, "down", "0")
		// outBlockState = strings.ReplaceAll(outBlockState, "up", "1")
		// outBlockState = strings.ReplaceAll(outBlockState, "north", "2")
		// outBlockState = strings.ReplaceAll(outBlockState, "south", "3")
		// outBlockState = strings.ReplaceAll(outBlockState, "west", "4")
		// outBlockState = strings.ReplaceAll(outBlockState, "east", "5")
	}
	if outBlockName == "wall_sign" {
		outBlockState = strings.ReplaceAll(outBlockState, "facing", "facing_direction")
		outBlockState = strings.ReplaceAll(outBlockState, "facing_direction_direction", "facing_direction")
		outBlockState = strings.ReplaceAll(outBlockState, "north", "2")
		outBlockState = strings.ReplaceAll(outBlockState, "south", "3")
		outBlockState = strings.ReplaceAll(outBlockState, "west", "4")
		outBlockState = strings.ReplaceAll(outBlockState, "east", "5")
	}
	//  else if outBlockName == "piston_extension" {
	// 	outBlockName = "piston_arm_collision"
	// 	outBlockState = "block_data=0"
	// } else if (outBlockName == "piston_head") || (outBlockName == "pistonArmCollision") {
	// 	outBlockName = "piston_arm_collision"
	// 	outBlockState = "block_data=0"
	// }
	// if strings.HasPrefix(inBlockName, "mangrove_propagule") {
	// 	if strings.HasPrefix(inBlockState, "hanging=0b") || strings.Contains(inBlockState, `hanging="false"`) {
	// 		outBlockName = "mangrove_propagule"
	// 	} else {
	// 		outBlockName = "mangrove_propagule_hanging"
	// 	}
	// 	trimedInState := inBlockState
	// 	if idx := strings.Index(trimedInState, "propagule_stage="); idx != -1 {
	// 		trimedInState = strings.TrimPrefix(trimedInState[idx:], `propagule_stage=`)
	// 		// inBlockState = inBlockState[len(inBlockState)-2 : len(inBlockState)-1]
	// 	} else if idx := strings.Index(trimedInState, "stage="); idx != -1 {
	// 		trimedInState = strings.TrimPrefix(trimedInState[idx:], `stage="`)
	// 		trimedInState = trimedInState[len(trimedInState)-2 : len(trimedInState)-1]
	// 	}

	// 	outBlockState = "facing_direction:0, growth:" + trimedInState
	// }
	// if strings.HasPrefix(inBlockName, "sculk_catalyst") {
	// 	outBlockName = "sculk_catalyst"
	// 	if strings.Contains(inBlockState, `bloom="true"`) || strings.Contains(inBlockState, `bloom=1b`) {
	// 		outBlockState = "bloom=1b"
	// 	} else if strings.Contains(inBlockState, `bloom="`) || strings.Contains(inBlockState, `bloom=0b`) {
	// 		outBlockState = "bloom=0b"
	// 	} else {
	// 		panic(inBlockState)
	// 	}
	// } else if strings.HasPrefix(inBlockName, "sculk_sensor") || strings.HasPrefix(inBlockName, "calibrated_sculk_sensor") {
	// 	outBlockName = "sculk_sensor"
	// 	if strings.Contains(inBlockState, `power="0"`) || strings.Contains(inBlockState, "powered_bit=0b") {
	// 		outBlockState = "sculk_sensor_phase=0"
	// 	} else if strings.Contains(inBlockState, `power="`) || strings.Contains(inBlockState, "powered_bit=1b") {
	// 		outBlockState = "sculk_sensor_phase=1"
	// 	} else {
	// 		outBlockState = "sculk_sensor_phase=0"
	// 	}
	// } else if strings.HasPrefix(inBlockName, "sculk_vein") {
	// 	outBlockName = "sculk_vein"
	// 	outBlockState = "multi_face_direction_bits=0"
	// }
	//  if strings.HasPrefix(inBlockName, "sculk_shrieker") {
	// 	outBlockName = "sculk_shrieker"
	// 	if strings.Contains(inBlockState, `shrieking="true"`) || strings.Contains(inBlockState, `active=1b`) {
	// 		outBlockState = "active=1b"
	// 	} else if strings.Contains(inBlockState, `shrieking="false"`) || strings.Contains(inBlockState, `active=0b`) {
	// 		outBlockState = "active=0b"
	// 	} else {
	// 		panic(inBlockState)
	// 	}
	// }
	// if strings.HasPrefix(inBlockName, "sculk") {
	// 	outBlockName = "sculk"
	// 	outBlockState = ""
	// }
	return inBlockName, inBlockState, outBlockName, outBlockState
}
