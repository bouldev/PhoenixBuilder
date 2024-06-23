package packet

const (
	IDLogin = iota + 1
	IDPlayStatus
	IDServerToClientHandshake
	IDClientToServerHandshake
	IDDisconnect
	IDResourcePacksInfo
	IDResourcePackStack
	IDResourcePackClientResponse
	IDText
	IDSetTime
	IDStartGame
	IDAddPlayer
	IDAddActor
	IDRemoveActor
	IDAddItemActor
	_
	IDTakeItemActor
	IDMoveActorAbsolute
	IDMovePlayer
	IDPassengerJump
	IDUpdateBlock
	IDAddPainting
	IDTickSync

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	//
	// Netease: new packet
	IDLevelSoundEventV1

	IDLevelEvent
	IDBlockEvent
	IDActorEvent
	IDMobEffect
	IDUpdateAttributes
	IDInventoryTransaction
	IDMobEquipment
	IDMobArmourEquipment
	IDInteract
	IDBlockPickRequest
	IDActorPickRequest
	IDPlayerAction
	_
	IDHurtArmour
	IDSetActorData
	IDSetActorMotion
	IDSetActorLink
	IDSetHealth
	IDSetSpawnPosition
	IDAnimate
	IDRespawn
	IDContainerOpen
	IDContainerClose
	IDPlayerHotBar
	IDInventoryContent
	IDInventorySlot
	IDContainerSetData
	IDCraftingData
	IDCraftingEvent
	IDGUIDataPickItem

	// PhoenixBuilder specific comments.
	// Author: Liliya233
	//
	// Netease: missing
	IDAdventureSettings

	IDBlockActorData
	IDPlayerInput
	IDLevelChunk
	IDSetCommandsEnabled
	IDSetDifficulty
	IDChangeDimension
	IDSetPlayerGameType
	IDPlayerList
	IDSimpleEvent
	IDEvent
	IDSpawnExperienceOrb
	IDClientBoundMapItemData
	IDMapInfoRequest
	IDRequestChunkRadius
	IDChunkRadiusUpdated
	IDItemFrameDropItem
	IDGameRulesChanged
	IDCamera
	IDBossEvent
	IDShowCredits
	IDAvailableCommands
	IDCommandRequest
	IDCommandBlockUpdate
	IDCommandOutput
	IDUpdateTrade
	IDUpdateEquip
	IDResourcePackDataInfo
	IDResourcePackChunkData
	IDResourcePackChunkRequest
	IDTransfer
	IDPlaySound
	IDStopSound
	IDSetTitle
	IDAddBehaviourTree
	IDStructureBlockUpdate
	IDShowStoreOffer
	IDPurchaseReceipt
	IDPlayerSkin
	IDSubClientLogin
	IDAutomationClientConnect
	IDSetLastHurtBy
	IDBookEdit
	IDNPCRequest
	IDPhotoTransfer
	IDModalFormRequest
	IDModalFormResponse
	IDServerSettingsRequest
	IDServerSettingsResponse
	IDShowProfile
	IDSetDefaultGameType
	IDRemoveObjective
	IDSetDisplayObjective
	IDSetScore
	IDLabTable
	IDUpdateBlockSynced
	IDMoveActorDelta
	IDSetScoreboardIdentity
	IDSetLocalPlayerAsInitialised
	IDUpdateSoftEnum
	IDNetworkStackLatency
	_

	// PhoenixBuilder specific comments.
	// Author: Liliya233
	//
	// Netease: missing
	IDScriptCustomEvent

	IDSpawnParticleEffect
	IDAvailableActorIdentifiers

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	//
	// Netease: new packet
	IDLevelSoundEventV2

	IDNetworkChunkPublisherUpdate
	IDBiomeDefinitionList
	IDLevelSoundEvent
	IDLevelEventGeneric
	IDLecternUpdate
	_
	IDAddEntity
	IDRemoveEntity
	IDClientCacheStatus

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	//
	// Netease: 131 -> 130
	IDOnScreenTextureAnimation
	// PhoenixBuilder specific changes.
	// Author: Liliya233
	//
	// Netease: 130 -> 131
	IDMapCreateLockedCopy

	IDStructureTemplateDataRequest
	IDStructureTemplateDataResponse
	_
	IDClientCacheBlobStatus
	IDClientCacheMissResponse
	IDEducationSettings
	IDEmote
	IDMultiPlayerSettings
	IDSettingsCommand
	IDAnvilDamage
	IDCompletedUsingItem
	IDNetworkSettings
	IDPlayerAuthInput
	IDCreativeContent
	IDPlayerEnchantOptions
	IDItemStackRequest
	IDItemStackResponse
	IDPlayerArmourDamage
	IDCodeBuilder
	IDUpdatePlayerGameType
	IDEmoteList
	IDPositionTrackingDBServerBroadcast
	IDPositionTrackingDBClientRequest
	IDDebugInfo
	IDPacketViolationWarning
	IDMotionPredictionHints
	IDAnimateEntity
	IDCameraShake
	IDPlayerFog
	IDCorrectPlayerMovePrediction
	IDItemComponent
	IDFilterText
	IDClientBoundDebugRenderer
	IDSyncActorProperty
	IDAddVolumeEntity
	IDRemoveVolumeEntity
	IDSimulationType
	IDNPCDialogue
	IDEducationResourceURI
	IDCreatePhoto
	IDUpdateSubChunkBlocks

	// PhoenixBuilder specific comments.
	// Author: Liliya233
	//
	// Netease: missing
	IDPhotoInfoRequest

	IDSubChunk
	IDSubChunkRequest
	IDClientStartItemCooldown
	IDScriptMessage
	IDCodeBuilderSource
	IDTickingAreasLoadStatus
	IDDimensionData
	IDAgentAction
	IDChangeMobProperty
	IDLessonProgress
	IDRequestAbility
	IDRequestPermissions
	IDToastRequest
	IDUpdateAbilities
	IDUpdateAdventureSettings
	IDDeathInfo
	IDEditorNetwork
	IDFeatureRegistry
	IDServerStats
	IDRequestNetworkSettings
	IDGameTestRequest
	IDGameTestResults
	IDUpdateClientInputLocks

	// PhoenixBuilder specific comments.
	// Author: Liliya233
	//
	// Netease: missing
	IDClientCheatAbility

	IDCameraPresets
	IDUnlockedRecipes

	// PhoenixBuilder specific changes.
	// Author: LNSSPsd, Liliya233, Happy2018new
	IDPyRpc

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	IDChangeModel              // Netease: new packet
	IDStoreBuySucc             // Netease: new packet
	IDNeteaseJson              // Netease: new packet
	IDChangeModelTexture       // Netease: new packet
	IDChangeModelOffset        // Netease: new packet
	IDChangeModelBind          // Netease: new packet
	IDHungerAttr               // Netease: new packet
	IDSetDimensionLocalTime    // Netease: new packet
	IDWithdrawFurnaceXp        // Netease: new packet
	IDSetDimensionLocalWeather // Netease: new packet

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	IDCustomV1             = iota + 13 // Netease: new packet
	IDCombine                          // Netease: new packet
	IDVConnection                      // Netease: new packet
	IDTransport                        // Netease: new packet
	IDCustomV2                         // Netease: new packet
	IDConfirmSkin                      // Netease: new packet
	IDTransportNoCompress              // Netease: new packet
	IDMobEffectV2                      // Netease: new packet
	IDMobBlockActorChanged             // Netease: new packet
	IDChangeActorMotion                // Netease: new packet
	IDAnimateEmoteEntity               // Netease: new packet

	// PhoenixBuilder specific changes.
	// Author: Liliya233
	IDCameraInstruction             = iota + 79 // Netease: 301 -> 300
	IDCompressedBiomeDefinitionList             // Netease: 302 -> 301
	IDTrimData                                  // Netease: 303 -> 302
	IDOpenSign                                  // Netease: 304 -> 303
	IDAgentAnimation                            // Netease: 305 -> 304
)
