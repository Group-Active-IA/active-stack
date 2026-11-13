package timing

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func iv(s, e int) Interval { return Interval{time.Duration(s), time.Duration(e)} }

func TestUnion_Examples(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []Interval
		want []Interval
		tot  time.Duration
	}{
		{"vacio", nil, nil, 0},
		{"uno", []Interval{iv(1, 4)}, []Interval{iv(1, 4)}, 3},
		{"disjuntos", []Interval{iv(5, 7), iv(1, 2)}, []Interval{iv(1, 2), iv(5, 7)}, 3},
		{"solapados", []Interval{iv(1, 5), iv(3, 8)}, []Interval{iv(1, 8)}, 7},
		{"contiguos", []Interval{iv(1, 3), iv(3, 6)}, []Interval{iv(1, 6)}, 5},
		{"contenido", []Interval{iv(1, 10), iv(2, 3)}, []Interval{iv(1, 10)}, 9},
		{"cadena fuera de orden", []Interval{iv(6, 9), iv(1, 4), iv(3, 7)}, []Interval{iv(1, 9)}, 8},
		{"vacios e invertidos se ignoran", []Interval{iv(4, 4), iv(9, 2), iv(1, 2)}, []Interval{iv(1, 2)}, 1},
	} {
		got := Union(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: Union = %v, want %v", tc.name, got, tc.want)
		}
		if tot := Total(tc.in); tot != tc.tot {
			t.Errorf("%s: Total = %v, want %v", tc.name, tot, tc.tot)
		}
	}
}

func TestUnion_DoesNotMutateInput(t *testing.T) {
	in := []Interval{iv(6, 9), iv(1, 4), iv(3, 7)}
	cp := append([]Interval(nil), in...)
	Union(in)
	if !reflect.DeepEqual(in, cp) {
		t.Errorf("Union modifico la entrada: %v, era %v", in, cp)
	}
}

// Propiedad: contra un oraculo independiente (celdas unitarias cubiertas) con
// entradas aleatorias y permutadas, el total no depende del orden y la union
// es ordenada, disjunta y sin intervalos contiguos.
func TestUnion_PropertyAgainstCellOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(20261111))
	for iter := 0; iter < 2000; iter++ {
		n := rng.Intn(8)
		var in []Interval
		for i := 0; i < n; i++ {
			s := rng.Intn(20)
			in = append(in, iv(s, s+rng.Intn(10)-1)) // incluye vacios (len 0) e invertidos
		}
		var cells [40]bool
		for _, x := range in {
			for c := x.Start; c < x.End; c++ {
				cells[c] = true
			}
		}
		var want time.Duration
		for _, c := range cells {
			if c {
				want++
			}
		}
		var prev []Interval
		for p := 0; p < 4; p++ {
			rng.Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
			got := Union(in)
			if tot := Total(in); tot != want {
				t.Fatalf("iter %d: Total(%v) = %v, oraculo %v", iter, in, tot, want)
			}
			var sum time.Duration
			for i, g := range got {
				sum += g.End - g.Start
				if g.End <= g.Start {
					t.Fatalf("iter %d: intervalo vacio en la union %v", iter, got)
				}
				if i > 0 && g.Start <= got[i-1].End {
					t.Fatalf("iter %d: union no disjunta ni separada: %v", iter, got)
				}
			}
			if sum != want {
				t.Fatalf("iter %d: suma de la union %v, oraculo %v (%v)", iter, sum, want, got)
			}
			if prev != nil && !reflect.DeepEqual(prev, got) {
				t.Fatalf("iter %d: la union depende del orden: %v vs %v", iter, prev, got)
			}
			prev = got
		}
	}
}
