import { IconArrowDown, IconArrowUp } from "@tabler/icons-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { FlashValue } from "@/components/flash-value"
import {
  Card,
  CardAction,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { TrafficTotalSummary } from "@/lib/api/traffic"
import { bytesToGBString } from "@/lib/units"

type SectionCardsProps = {
  total1h?: TrafficTotalSummary
  total24h?: TrafficTotalSummary
  totalAll?: TrafficTotalSummary
  isLoading?: boolean
}

export function SectionCards(props: SectionCardsProps) {
  const { t } = useTranslation()
  const up1 = props.total1h?.upload ?? 0
  const down1 = props.total1h?.download ?? 0
  const up24 = props.total24h?.upload ?? 0
  const down24 = props.total24h?.download ?? 0
  const upAll = props.totalAll?.upload ?? 0
  const downAll = props.totalAll?.download ?? 0

  return (
    <div className="grid grid-cols-1 gap-4 px-4 lg:px-6 @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
      <Card className="@container/card">
        <CardHeader>
          <CardDescription>{t("dashboard.uplinkLast1h")}</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            {props.isLoading ? "-" : <FlashValue value={`${bytesToGBString(up1)} GB`} />}
          </CardTitle>
          <CardAction>
            <Badge variant="outline">
              <IconArrowUp />
              {t("dashboard.uplink")}
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1.5 text-sm">
          <div className="text-muted-foreground">
            {t("dashboard.allTimeUplink", { value: `${bytesToGBString(upAll)} GB` })}
          </div>
        </CardFooter>
      </Card>
      <Card className="@container/card">
        <CardHeader>
          <CardDescription>{t("dashboard.downlinkLast1h")}</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            {props.isLoading ? "-" : <FlashValue value={`${bytesToGBString(down1)} GB`} />}
          </CardTitle>
          <CardAction>
            <Badge variant="outline">
              <IconArrowDown />
              {t("dashboard.downlink")}
            </Badge>
          </CardAction>
        </CardHeader>
        <CardFooter className="flex-col items-start gap-1.5 text-sm">
          <div className="text-muted-foreground">
            {t("dashboard.allTimeDownlink", { value: `${bytesToGBString(downAll)} GB` })}
          </div>
        </CardFooter>
      </Card>
      <Card className="@container/card">
        <CardHeader>
          <CardDescription>{t("dashboard.uplinkLast24h")}</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            {props.isLoading ? "-" : <FlashValue value={`${bytesToGBString(up24)} GB`} />}
          </CardTitle>
          <CardAction>
            <Badge variant="outline">
              <IconArrowUp />
              {t("dashboard.uplink")}
            </Badge>
          </CardAction>
        </CardHeader>
      </Card>
      <Card className="@container/card">
        <CardHeader>
          <CardDescription>{t("dashboard.downlinkLast24h")}</CardDescription>
          <CardTitle className="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
            {props.isLoading ? "-" : <FlashValue value={`${bytesToGBString(down24)} GB`} />}
          </CardTitle>
          <CardAction>
            <Badge variant="outline">
              <IconArrowDown />
              {t("dashboard.downlink")}
            </Badge>
          </CardAction>
        </CardHeader>
      </Card>
    </div>
  )
}
