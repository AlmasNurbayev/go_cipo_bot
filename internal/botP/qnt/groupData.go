package qnt

import (
	"slices"

	"github.com/AlmasNurbayev/go_cipo_bot/internal/models"
)

type GroupDataI struct {
	Goods  Goods
	Stores []StoreEntity
}

type Goods struct {
	NomVids []NomVidsEntity
	Sum     float32
	Qnt     int64
}

type NomVidsEntity struct {
	Name          string
	ProductGroups []ProductGroupsEntity
	Sum           float32
	Qnt           int64
}

type ProductGroupsEntity struct {
	Name       string
	SizeGroups []SizeGroupsEntity
	Sum        float32
	Qnt        int64
}

type SizeGroupsEntity struct {
	Name  string
	Begin string
	End   string
	Sum   float32
	Qnt   int64
}

type StoreEntity struct {
	Name string
	Sum  float32
	Qnt  int64
}

func transformQntData(data []models.ProductOnlyQnt) GroupDataI {
	var result GroupDataI

	stores := []StoreEntity{}
	Goods := Goods{}
	for _, item := range data {

		// группируем по ном видам
		nomVidName := item.Nom_vid.String
		if nomVidName == "" {
			nomVidName = "неизвестно"
		}
		if !slices.ContainsFunc(Goods.NomVids, func(nv NomVidsEntity) bool { return nv.Name == nomVidName }) {
			Goods.NomVids = append(Goods.NomVids, NomVidsEntity{Name: nomVidName, Sum: item.Sum, Qnt: item.Qnt})
		} else {
			for i := range Goods.NomVids {
				if Goods.NomVids[i].Name == nomVidName {
					Goods.NomVids[i].Sum += item.Sum_zakup * float32(item.Qnt)
					Goods.NomVids[i].Qnt += item.Qnt
				}
			}
		}

		// группируем по продукт группам
		nomVidIndex := slices.IndexFunc(Goods.NomVids, func(nv NomVidsEntity) bool { return nv.Name == nomVidName })
		if nomVidIndex != -1 {
			prodGroupName := item.Product_group_name
			if prodGroupName == "" {
				prodGroupName = "неизвестно"
			}
			if !slices.ContainsFunc(Goods.NomVids[nomVidIndex].ProductGroups, func(pg ProductGroupsEntity) bool { return pg.Name == prodGroupName }) {
				Goods.NomVids[nomVidIndex].ProductGroups = append(Goods.NomVids[nomVidIndex].ProductGroups, ProductGroupsEntity{Name: prodGroupName, Sum: item.Sum, Qnt: item.Qnt})
			} else {
				for j := range Goods.NomVids[nomVidIndex].ProductGroups {
					if Goods.NomVids[nomVidIndex].ProductGroups[j].Name == prodGroupName {
						Goods.NomVids[nomVidIndex].ProductGroups[j].Sum += item.Sum_zakup * float32(item.Qnt)
						Goods.NomVids[nomVidIndex].ProductGroups[j].Qnt += item.Qnt
					}
				}
			}
		}

		// группируем по size группам
		if nomVidIndex != -1 {
			prodGroupName := item.Product_group_name
			prodGroupIndex := slices.IndexFunc(Goods.NomVids[nomVidIndex].ProductGroups, func(pg ProductGroupsEntity) bool { return pg.Name == prodGroupName })
			if prodGroupIndex != -1 {
				sizeGroupName := item.Size_name
				if sizeGroupName == "" {
					sizeGroupName = "неизвестно"
				}
				if !slices.ContainsFunc(Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups, func(sg SizeGroupsEntity) bool { return sg.Name == sizeGroupName }) {
					Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups = append(Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups, SizeGroupsEntity{
						Name: sizeGroupName,
						//Begin: item.Size_begin.String,
						//End:   item.Size_end.String,
						Sum: item.Sum,
						Qnt: item.Qnt,
					})
				} else {
					for k := range Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups {
						if Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups[k].Name == sizeGroupName {
							Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups[k].Sum += item.Sum_zakup * float32(item.Qnt)
							Goods.NomVids[nomVidIndex].ProductGroups[prodGroupIndex].SizeGroups[k].Qnt += item.Qnt
						}
					}
				}
			}
		}

		// группируем по складам
		storeName := item.Store_name
		if storeName == "" {
			storeName = "неизвестно"
		}
		if !slices.ContainsFunc(stores, func(s StoreEntity) bool { return s.Name == storeName }) {
			stores = append(stores, StoreEntity{Name: storeName, Sum: item.Sum, Qnt: item.Qnt})
		} else {
			for i := range stores {
				if stores[i].Name == storeName {
					stores[i].Sum += item.Sum_zakup * float32(item.Qnt)
					stores[i].Qnt += item.Qnt
				}
			}
		}
	}

	result.Stores = stores
	result.Goods = Goods

	return result

}
