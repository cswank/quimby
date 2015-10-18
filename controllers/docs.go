package controllers

//Options produces documentation for the API
func Options(args *Args) error {
	if args.Gadget != nil {
		return gadgetOptions(args)
	}
	return gadgetsOptions(args)

}

func gadgetsOptions(args *Args) error {
	args.W.Write([]byte("{}"))
	return nil
}

func gadgetOptions(args *Args) error {
	args.W.Write([]byte("{}"))
	return nil
}
