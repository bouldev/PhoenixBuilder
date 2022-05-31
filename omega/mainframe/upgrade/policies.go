package upgrade

func Upgrade() {
	storageRoot := Policy_1()
	Policy_2(storageRoot)
	Policy_3(storageRoot)
	Policy_4(storageRoot)
}
