package beater

func List2Set (ss []string) ( mm map[string]struct{}) {
    //converts lists to sort of sets, to avoid having set literals in config
    
    mm = make ( map[string]struct{}, len(ss) )
    for _, s := range ss {
        mm[s] = struct{}{}
    }
    return mm
}

