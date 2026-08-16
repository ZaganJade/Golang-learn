package main

import "fmt"

func swap(a, b *int) {
	*a, *b = *b, *a
}

func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

func swapByValue(a, b int) {
	a, b = b, a
}

func updateSliceByValue(s []string, newItem string) {
	s = append(s, newItem)
}

func main() {
	fmt.Println("== swap ==")

	x, y := 5, 10
	fmt.Println("Sebelum swapByValue:", x, y)
	swapByValue(x, y)
	fmt.Println("Setelah swapByValue:", x, y, "(enih By Value)")

	a, b := 5, 10
	fmt.Println("\nSebelum swap (pointer):", a, b)
	swap(&a, &b)
	fmt.Println("Setelah swap (pointer):", a, b, "(enih By Pointer)")

	fmt.Println("\n== updateSlice ==")

	items1 := []string{"pensil", "buku"}
	fmt.Println("Sebelum updateSliceByValue:", items1)
	updateSliceByValue(items1, "penghapus")
	fmt.Println("Setelah updateSliceByValue:", items1, "(tidak berubah)")

	items2 := []string{"pensil", "buku"}
	fmt.Println("\nSebelum updateSlice (pointer):", items2)
	updateSlice(&items2, "penghapus")
	fmt.Println("Setelah updateSlice (pointer):", items2, "(berubah)")
}