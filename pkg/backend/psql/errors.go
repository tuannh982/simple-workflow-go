package psql

func HandleSQLError(err error) error {
	//if err != nil {
	//	if strings.Contains(err.Error(), "could not serialize access due to concurrent") {
	//		return worker.SuppressedError
	//	}
	//}
	return err
}
