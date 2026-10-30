import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import { getErrorToast } from "@/utilities";

export default {
    data() {
        const endDate = new Date();
        const startDate = new Date(endDate);
        startDate.setDate(startDate.getDate() - 6);
        startDate.setHours(0, 0, 0, 0);
        return {
            startDate,
            endDate,
            todayDate: [
                endDate.getFullYear(),
                String(endDate.getMonth() + 1).padStart(2, "0"),
                String(endDate.getDate()).padStart(2, "0"),
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
            const startDate = this.parseDateRangeValue(range.startDate);
            const selectedEnd = this.parseDateRangeValue(range.endDate);
            if (!startDate || !selectedEnd) {
                this.$store.dispatch("createToast", getErrorToast("Select a valid analytics date range."));
                return;
            }
            startDate.setHours(0, 0, 0, 0);
            selectedEnd.setHours(23, 59, 59, 999);
            const now = new Date();
            const endDate = selectedEnd > now ? now : selectedEnd;
            if (startDate >= endDate) {
                this.$store.dispatch("createToast", getErrorToast("Select a valid analytics date range."));
                return;
            }
            this.startDate = startDate;
            this.endDate = endDate;
            this.fetch();
        },
        parseDateRangeValue(value) {
            const date = value instanceof Date ? new Date(value) : new Date(`${value}T00:00:00`);
            return Number.isNaN(date.getTime()) ? null : date;
        },
    },
};
