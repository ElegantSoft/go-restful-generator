import * as React from "react"
import { Card, CardHeader, CardContent } from "@material-ui/core"
import { makeStyles } from "@material-ui/core/styles"
import {
  DateField,
  EditButton,
  NumberField,
  TextField,
  BooleanField,
  useTranslate,
  RecordMap,
  Identifier,
  Record,
} from "react-admin"

const useListStyles = makeStyles((theme) => ({
  card: {
    height: "100%",
    display: "flex",
    flexDirection: "column",
    margin: "0.5rem 0",
  },
  cardTitleContent: {
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  cardContent: theme.typography.body1,
  cardContentRow: {
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
    margin: "0.5rem 0",
  },
}))

interface MobileGridProps {
  ids?: Identifier[]
  data?: RecordMap<Record>
  basePath?: string
}

export const MobileGrid = (props: MobileGridProps) => {
  const { ids, data, basePath } = props
  const translate = useTranslate()
  const classes = useListStyles()

  if (!ids || !data || !basePath) {
    return null
  }

  return (
    <div style={{ margin: "1em" }}>
      {ids.map((id) => (
        <Card key={id} className={classes.card}>
          <CardHeader
            title={
              <div className={classes.cardTitleContent}>
                <span>
                  المعرف: <TextField record={data[id]} source="id" />
                </span>
                <EditButton
                  resource="commands"
                  basePath={basePath}
                  record={data[id]}
                />
              </div>
            }
          />
          <CardContent className={classes.cardContent}>
            <span className={classes.cardContentRow}>
              {"الإسم: "} <TextField record={data[id]} source="name" />
            </span>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
