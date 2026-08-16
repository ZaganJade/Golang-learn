package main

import "fmt"

func main() {
	// bikin deklarasi dengan tipe data yang berbeda
	var nama string = "Aku Golang"
	var umur int = 20
	var ipk float64 = 3.75
	var aktif bool = true
	var hobi []string = []string{"Mengodonf", "Mengodonf EyAy", "Explore tech", "Nak nonton anime"}

	fmt.Println("=== Variabel ===")
	fmt.Println("Nama :", nama)
	fmt.Println("Umur :", umur)
	fmt.Println("IPK  :", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Hobi :", hobi)

	// Map buat menyimpan data mahasiswa dgn nama sebagai key dan nilai sebagai value
	mahasiswa := make(map[string]int)

	// kita kamsih datanya ke map
	mahasiswa["Rusdi"] = 85
	mahasiswa["Amba"] = 90
	mahasiswa["Ihya"] = 88

	fmt.Println("\n=== Data Mahasiswa Setelah Ditambah ===")
	fmt.Println(mahasiswa)

	// kimta cari data dari Rusdi Yagesyak
	nilai, ada := mahasiswa["Rusdi"]
	if ada {
		fmt.Println("\nNilai Rusdi:", nilai)
	} else {
		fmt.Println("\nData Rusdi tidak ditemukan")
	}

	// Kimta hapus data Amba dari map yagesyak
	delete(mahasiswa, "Amba")

	fmt.Println("\n=== Data Mahasiswa Setelah Amba Dihapus ===")
	fmt.Println(mahasiswa)

	// mencari seluruh isi map
	fmt.Println("\n=== Seluruh Data Mahasiswa ===")
	for nama, nilai := range mahasiswa {
		fmt.Printf("%s : %d\n", nama, nilai)
	}
}

