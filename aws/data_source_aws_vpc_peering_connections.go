package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsVpcPeeringConnections() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsVpcPeeringConnectionsRead,
		Schema: map[string]*schema.Schema{
			"filter": ec2CustomFiltersSchema(),

			"tags": tagsSchemaComputed(),

			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsVpcPeeringConnectionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[DEBUG] Reading VPC Peering Connections.")

	req := &ec2.DescribeVpcPeeringConnectionsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.VpcPeeringConnectionIds = aws.StringSlice([]string{id.(string)})
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"status-code": d.Get("status").(string),
		},
	)

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] DescribeVpcPeeringConnections %s\n", req)
	resp, err := conn.DescribeVpcPeeringConnections(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.VpcPeeringConnections) == 0 {
		return fmt.Errorf("no matching VpcPeeringConnections found")
	}

	pcxs := make([]string, 0)

	for _, pcx := range resp.VpcPeeringConnections {
		pcxs = append(pcxs, aws.StringValue(pcx.VpcPeeringConnectionId))
	}

	if len(pcxs) < 1 {
		return fmt.Errorf("less than 1 VpcPeeringConnectionId has returned")
	}

	d.SetId(time.Now().UTC().String())
	if err := d.Set("ids", pcxs); err != nil {
		return fmt.Errorf("Error setting VpcPeeringConnections ids: %s", err)
	}

	return nil
}
