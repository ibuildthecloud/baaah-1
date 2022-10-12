package router

import (
	"k8s.io/utils/strings/slices"
)

type FinalizerHandler struct {
	FinalizerID string
	Next        Handler
}

func (f FinalizerHandler) Handle(req Request, resp Response) error {
	obj := req.Object
	if obj == nil {
		return nil
	}

	if obj.GetDeletionTimestamp().IsZero() {
		if !slices.Contains(obj.GetFinalizers(), f.FinalizerID) {
			obj.SetFinalizers(append(obj.GetFinalizers(), f.FinalizerID))
			if err := req.Client.Update(req.Ctx, obj); err != nil {
				return err
			}
			resp.Objects(obj)
		}
		return nil
	}

	if !slices.Contains(obj.GetFinalizers(), f.FinalizerID) {
		return nil
	}

	if err := f.Next.Handle(req, resp); err != nil {
		return err
	}

	if !obj.GetDeletionTimestamp().IsZero() {
		if slices.Contains(obj.GetFinalizers(), f.FinalizerID) {
			ff := obj.GetFinalizers()
			for i := 0; i < len(ff); i++ {
				if ff[i] == f.FinalizerID {
					ff = append(ff[:i], ff[i+1:]...)
					i--
				}
			}
			obj.SetFinalizers(ff)

			if err := req.Client.Update(req.Ctx, obj); err != nil {
				return err
			}

			resp.Objects(obj)
		}
	}

	return nil
}
