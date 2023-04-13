package pagerduty

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/heimweh/go-pagerduty/pagerduty"
)

func dataSourcePagerDutySchedule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePagerDutyScheduleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_ids": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourcePagerDutyScheduleRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(*Config).Client()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Reading PagerDuty schedule")

	searchName := d.Get("name").(string)

	o := &pagerduty.ListSchedulesOptions{
		Query: searchName,
	}

	o2 := &pagerduty.GetScheduleOptions{}

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, _, err := client.Schedules.List(o)
		if err != nil {
			// Delaying retry by 30s as recommended by PagerDuty
			// https://developer.pagerduty.com/docs/rest-api-v2/rate-limiting/#what-are-possible-workarounds-to-the-events-api-rate-limit
			time.Sleep(30 * time.Second)
			return resource.RetryableError(err)
		}

		var found *pagerduty.Schedule

		for _, schedule := range resp.Schedules {
			if schedule.Name == searchName {
				found = schedule
				break
			}
		}

		resp2, _, err := client.Schedules.Get(found.ID, o2)

		var user_ids string
		for _, users := range resp2.ScheduleLayers[0].Users {
			user_ids = strings.Join([]string{user_ids, users.User.ID}, ",")
		}

		if found == nil {
			return resource.NonRetryableError(
				fmt.Errorf("Unable to locate any schedule with the name: %s", searchName),
			)
		}

		d.SetId(found.ID)
		d.Set("name", found.Name)
		d.Set("user_ids", strings.Trim(user_ids, ","))

		return nil
	})
}
