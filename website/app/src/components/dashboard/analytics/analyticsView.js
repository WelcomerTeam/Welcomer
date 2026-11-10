import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import { getErrorToast } from "@/utilities";

function getDateRangeEnd(date, now) {
    const endOfSelectedDate = new Date(date);
    endOfSelectedDate.setHours(23, 59, 59, 999);

    return endOfSelectedDate > now ? new Date(now) : endOfSelectedDate;
}

export default {
    inject: ["analyticsDateRange"],
    data() {
        const now = new Date();
        const endDate = new Date(now);
        const startDate = new Date(endDate);
        startDate.setDate(startDate.getDate() - 6);
        startDate.setHours(0, 0, 0, 0);
        const selectedStartDate = new Date(`${this.analyticsDateRange.startDate}T00:00:00`);
        const selectedEndDate = new Date(`${this.analyticsDateRange.endDate}T00:00:00`);
        if (!Number.isNaN(selectedStartDate.getTime()) && !Number.isNaN(selectedEndDate.getTime())) {
            startDate.setTime(selectedStartDate.getTime());
            endDate.setTime(getDateRangeEnd(selectedEndDate, now).getTime());
        }
        return {
            startDate,
            endDate,
            todayDate: [
                now.getFullYear(),
                String(now.getMonth() + 1).padStart(2, "0"),
                String(now.getDate()).padStart(2, "0"),
            ].join("-"),
            isDataFetched: false,
            isDataError: false,
            data: null,
            requestID: 0,
        };
    },
    mounted() {
        this.fetch();
    },
    watch: {
        "$route.params.guildID"() {
            this.fetch();
        },
    },
    methods: {
        fetch() {
            this.isDataFetched = false;
            this.isDataError = false;
            const requestID = ++this.requestID;
            const query = new URLSearchParams({
                from: this.startDate.toISOString(),
                to: this.endDate.toISOString(),
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC",
            });
            dashboardAPI.getConfig(
                `${endpoints.EndpointGuildAnalytics(this.$route.params.guildID, this.analyticsEndpoint)}?${query}`,
                ({ config }) => {
                    if (requestID !== this.requestID) return;
                    if (!config) {
                        this.$store.dispatch("createToast", getErrorToast("You do not have permission to view this guild's analytics."));
                        this.isDataFetched = true;
                        this.isDataError = true;
                        return;
                    }
                    this.data = config;
                    this.isDataFetched = true;
                },
                (error) => {
                    if (requestID !== this.requestID) return;
                    this.$store.dispatch("createToast", getErrorToast(error));
                    this.isDataFetched = true;
                    this.isDataError = true;
                }
            );
        },
        handleDateRangeUpdate(range) {
            if (!this.$store.getters.guildHasWelcomerPro &&
                ["last_month", "custom"].includes(range.preset)) {
                return;
            }

            const startDate = this.parseDateRangeValue(range.startDate);
            const selectedEnd = this.parseDateRangeValue(range.endDate);
            if (!startDate || !selectedEnd) {
                this.$store.dispatch("createToast", getErrorToast("Select a valid analytics date range."));
                return;
            }
            startDate.setHours(0, 0, 0, 0);
            const now = new Date();
            const endDate = getDateRangeEnd(selectedEnd, now);
            if (startDate >= endDate) {
                this.$store.dispatch("createToast", getErrorToast("Select a valid analytics date range."));
                return;
            }
            this.startDate = startDate;
            this.endDate = endDate;
            Object.assign(this.analyticsDateRange, {
                preset: range.preset || "custom",
                startDate: range.startDate,
                endDate: range.endDate,
            });
            this.fetch();
        },
        parseDateRangeValue(value) {
            const date = value instanceof Date ? new Date(value) : new Date(`${value}T00:00:00`);
            return Number.isNaN(date.getTime()) ? null : date;
        },
    },
};
