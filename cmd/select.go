package cmd

import (
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"encoding/json"
	"github.com/spf13/cobra"
	"log"
)

func newSelectCmd() *cobra.Command {
	type Options struct {
		OptProject   string `validate:"required"`
		OptKind      string `validate:"required"`
		OptNamespace string
	}

	var (
		o = &Options{}
	)

	cmd := &cobra.Command{
		Use:   "select",
		Short: "Select Entity",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := datastore.NewClient(ctx, o.OptProject)
			if err != nil {
				log.Fatalf("Could not create datastore client: %v", err)
				return err
			}

			query := datastore.NewQuery(o.OptKind)
			if o.OptNamespace != "" {
				query = query.Namespace(o.OptNamespace)
			}

			var entitys []Entity

			keys, err := client.GetAll(ctx, query, &entitys)
			if err != nil {
				log.Fatalf("Could not Get Keys: %v", err)
				return err
			}

			for i, key := range keys {
				entity := entitys[i]
				entity.Props["__key__"] = key

				j, _ := json.Marshal(entity.Props)
				var ij bytes.Buffer
				json.Indent(&ij, j, "", "  ")
				js := ij.String()
				if len(keys) > i+1 {
					js += ","
				}
				cmd.Printf("%s\n", js)
			}
			cmd.Print("-----------------\n")
			cmd.Printf("count: %d \n", len(keys))
			cmd.Print("-----------------\n")

			return nil
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.Flags().StringVarP(&o.OptProject, "project", "p", "", "datastore project id [required]")
	cmd.Flags().StringVarP(&o.OptKind, "kind", "k", "", "datastore kind [required]")
	cmd.Flags().StringVarP(&o.OptNamespace, "namespace", "n", "", "datastore namespace")

	cmd.MarkFlagRequired("project")
	cmd.MarkFlagRequired("kind")

	return cmd
}

func init() {
	RootCmd.AddCommand(newSelectCmd())
}

type Entity struct {
	Props map[string]interface{}
}

func (e *Entity) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(e, ps)

	if fmerr, ok := err.(*datastore.ErrFieldMismatch); ok && fmerr != nil && fmerr.Reason == "no such struct field" {
	} else if err != nil {
		return err
	}

	e.Props = map[string]interface{}{}
	for _, p := range ps {
		e.Props[p.Name] = p.Value
	}

	return nil
}

func (e *Entity) Save() ([]datastore.Property, error) {
	pr, err := datastore.SaveStruct(e)
	if err != nil {
		return nil, err
	}
	return pr, nil
}
