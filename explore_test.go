package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestExplorer(t *testing.T) {
	tests := []struct {
		name    string
		start   int
		stride  int
		samples int
		want    []try
	}{
		{
			name:   "uniform sequence",
			start:  0,
			stride: 1,
			// samples: 9,
			want: []try{
				{0, 0},
				// ---
				{0, 1},
				{1, 0},
				{1, 1},
				// ---
				{0, 2},
				{1, 2},
				{2, 0},
				{2, 1},
				{2, 2},
				// ---
				{0, 3},
				{1, 3},
				{2, 3},
				{3, 0},
				{3, 1},
				{3, 2},
				{3, 3},
				// ---
				{0, 4},
				{1, 4},
				{2, 4},
				{3, 4},
				{4, 0},
				{4, 1},
				{4, 2},
				{4, 3},
				{4, 4},
				// ---
				{0, 5},
				// ...
			},
		},
		{
			name:   "stride sequence",
			start:  0,
			stride: 2,
			want: []try{
				{0, 0},
				// ---
				{0, 2},
				{1, 2},
				{2, 0},
				{2, 1},
				{2, 2},
				// ---
				{0, 4},
				{1, 4},
				{2, 4},
				{3, 4},
				{4, 0},
				{4, 1},
				{4, 2},
				{4, 3},
				{4, 4},
				// ---
				{0, 6},
				// ...
			},
		},
		{
			name:   "offset stride sequence",
			start:  1,
			stride: 2,
			want: []try{
				// ---
				{0, 1},
				{1, 0},
				{1, 1},
				// ---
				{0, 3},
				{1, 3},
				{2, 3},
				{3, 0},
				{3, 1},
				{3, 2},
				{3, 3},
				// ---
				{0, 5},
				// ...
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // run in parallel to validate thread safety

			exploreFunc := Explorer(tt.start, tt.stride)
			results := make([]try, 0, tt.samples)
			for i := 0; i < len(tt.want); i++ {
				results = append(results, exploreFunc())
			}

			if !reflect.DeepEqual(results, tt.want) {
				t.Errorf("ExploreGen() = %v, want %v", results, tt.want)
			}
		})
	}
}

func Test_splitExplore(t *testing.T) {
	tests := []struct {
		n    int
		want []exploreFunc // equivalent exploreFunc that yields same sequence
	}{
		{1, []exploreFunc{Explorer(0, 1)}},
		{2, []exploreFunc{Explorer(0, 2), Explorer(1, 2)}},
		{3, []exploreFunc{Explorer(0, 3), Explorer(1, 3), Explorer(2, 3)}},
		{-1, []exploreFunc{}}, // dont panic on negative ints
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			got := splitExplore(tt.n)

			if len(got) != len(tt.want) {
				t.Fatalf("len(splitExplore(%v)) = %d,  want %d", tt.n, len(got), len(tt.want))
			}

			for i := 0; i < len(got); i++ {
				compareGenerators(t, got[i], tt.want[i], 10)
			}
		})
	}
}

func compareGenerators(t *testing.T, f1, f2 exploreFunc, samples int) {
	t.Helper()

	r1 := make([]try, 0, samples)
	r2 := make([]try, 0, samples)
	for i := 0; i < samples; i++ {
		r1 = append(r1, f1())
		r2 = append(r2, f2())
	}

	t.Log("\nf1:", r1, "\nf2:", r2)
	if !reflect.DeepEqual(r1, r2) {
		t.Errorf("generated divergent seqs:\n\tf1: %v\n\tf2: %v", r1, r2)
	}
}

func BenchmarkGeneratorFunc(b *testing.B) {
	f := Explorer(0, 1)
	for i := 0; i < b.N; i++ {
		_ = f()
	}
}

/*
// Legacy boilerplate from when I was generating initial sample text (before
// massaging it in asciiflow)

func ExampleExploreGen() {
	f1, f2, f3, f4 := ExploreGen(0, 4), ExploreGen(1, 4), ExploreGen(2, 4), ExploreGen(3, 4)

    w := new(tabwriter.Writer)
    // minwidth, tabwidth, padding, padchar, flags
    w.Init(os.Stdout, 12, 8, 0, '\t', 0)
    defer w.Flush()

    fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "f(0,4)", "f(1,4)", "f(2,4)", "f(3,4)")
    fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "------", "------", "------", "------")
    for i := 0; i < 20; i++ {
        fmt.Fprintf(w, "\n %v\t%v\t%v\t%v\t", f1(), f2(), f3(), f4())
    }
}
*/
