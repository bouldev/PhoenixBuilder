package protocol

const (
	ContainerAnvilInput = iota
	ContainerAnvilMaterial
	ContainerAnvilResultPreview
	ContainerSmithingTableInput
	ContainerSmithingTableMaterial
	ContainerSmithingTableResultPreview
	ContainerArmor

	// PhoenixBuilder specific changes.
	// Author: Liliya233, Happy2018new
	//
	// container_items, and now we can sure
	// that it is at least used by agent robot
	ContainerLevelEntity

	ContainerBeaconPayment
	ContainerBrewingStandInput
	ContainerBrewingStandResult
	ContainerBrewingStandFuel
	ContainerCombinedHotBarAndInventory
	ContainerCraftingInput
	ContainerCraftingOutputPreview
	ContainerRecipeConstruction
	ContainerRecipeNature

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	//
	// Netease
	ContainerRecipeCustom

	ContainerRecipeItems
	ContainerRecipeSearch
	ContainerRecipeSearchBar
	ContainerRecipeEquipment
	ContainerRecipeBook
	ContainerEnchantingInput
	ContainerEnchantingMaterial
	ContainerFurnaceFuel
	ContainerFurnaceIngredient
	ContainerFurnaceResult
	ContainerHorseEquip
	ContainerHotBar
	ContainerInventory
	ContainerShulkerBox
	ContainerTradeIngredientOne
	ContainerTradeIngredientTwo
	ContainerTradeResultPreview
	ContainerOffhand
	ContainerCompoundCreatorInput
	ContainerCompoundCreatorOutputPreview
	ContainerElementConstructorOutputPreview
	ContainerMaterialReducerInput
	ContainerMaterialReducerOutput
	ContainerLabTableInput
	ContainerLoomInput
	ContainerLoomDye
	ContainerLoomMaterial
	ContainerLoomResultPreview
	ContainerBlastFurnaceIngredient
	ContainerSmokerIngredient
	ContainerTradeTwoIngredientOne
	ContainerTradeTwoIngredientTwo
	ContainerTradeTwoResultPreview
	ContainerGrindstoneInput
	ContainerGrindstoneAdditional
	ContainerGrindstoneResultPreview
	ContainerStonecutterInput
	ContainerStonecutterResultPreview
	ContainerCartographyInput
	ContainerCartographyAdditional
	ContainerCartographyResultPreview
	ContainerBarrel
	ContainerCursor
	ContainerCreatedOutput
	ContainerSmithingTableTemplate
)

const (
	ContainerTypeInventory = iota - 1
	ContainerTypeContainer
	ContainerTypeWorkbench
	ContainerTypeFurnace
	ContainerTypeEnchantment
	ContainerTypeBrewingStand
	ContainerTypeAnvil
	ContainerTypeDispenser
	ContainerTypeDropper
	ContainerTypeHopper
	ContainerTypeCauldron
	ContainerTypeCartChest
	ContainerTypeCartHopper
	ContainerTypeHorse
	ContainerTypeBeacon
	ContainerTypeStructureEditor
	ContainerTypeTrade
	ContainerTypeCommandBlock
	ContainerTypeJukebox
	ContainerTypeArmour
	ContainerTypeHand
	ContainerTypeCompoundCreator
	ContainerTypeElementConstructor
	ContainerTypeMaterialReducer
	ContainerTypeLabTable
	ContainerTypeLoom
	ContainerTypeLectern
	ContainerTypeGrindstone
	ContainerTypeBlastFurnace
	ContainerTypeSmoker
	ContainerTypeStonecutter
	ContainerTypeCartography
	ContainerTypeHUD
	ContainerTypeJigsawEditor
	ContainerTypeSmithingTable
	ContainerTypeChestBoat
)
