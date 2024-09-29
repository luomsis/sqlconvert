package sqlconvert

// SQL dialect types
type SQLDialect int

const (
	SQLServer     SQLDialect = iota + 1 // 1
	SQLOracle                           // 2
	SQLDB2                              // 3
	SQLMySQL                            // 4
	SQLPostgreSQL                       // 5
	SQLSybase                           // 6
	SQLInformix                         // 7
	SQLGreenplum                        // 8
	SQLSybaseASA                        // 9
	SQLTeradata                         // 10
	SQLNetezza                          // 11
	SQLMariaDB                          // 12
	SQLHive                             // 13
	SQLRedshift                         // 14
	SQLEsgynDB                          // 15
	SQLSybaseADS                        // 16
	SQLMariaDBORA                       // 17
)

// SQL clause scope
type SQLClauseScope int

const (
	SQLScopeAssignmentRightSide SQLClauseScope = iota + 1 // 1
	SQLScopeCaseFunc                                      // 2
	SQLScopeCursorParams                                  // 3
	SQLScopeFuncParams                                    // 4
	SQLScopeInsertValues                                  // 5
	SQLScopeProcParams                                    // 6
	SQLScopeSelectStmt                                    // 7
	SQLScopeTabCols                                       // 8
	SQLScopeTrgWhenCondition                              // 9
	SQLScopeVarDecl                                       // 10
	SQLScopeXMLSerializeFunc                              // 11
	SQLScopeSPAddType                                     // 12
	SQLScopeConvertFunc                                   // 13
	SQLScopeCastFunc                                      // 14
	SQLScopeObjTypeDecl                                   // 15
	SQLScopeFuncReturnDecl                                // 16
)
