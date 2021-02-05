import React from 'react';
import Container from 'react-bootstrap/Container';
import Row from 'react-bootstrap/Row';
import Col from 'react-bootstrap/Col';
import Table from 'react-bootstrap/Table';
import ButtonGroup from 'react-bootstrap/ButtonGroup';
import Button from 'react-bootstrap/Button';
import Moment from 'react-moment';
import moment from 'moment';
import {Link} from "react-router-dom";
import 'moment-duration-format';
import 'bootstrap/dist/css/bootstrap.min.css';
import './Overview.css';

function OverviewBanner(props) {
    return (
        <Container fluid>
            <div className="text-center"><h4>Current Host</h4></div>
            <br />

            <Container fluid className="text-white-50">
                <Row>
                    <Col sm={4} className="text-center">
                        <b>Hostname</b><br />{props.currentHost.hostname}<br /><br />
                    </Col>
                    <Col sm={4} className="text-center">
                        <b>Host Uptime</b>
                        <br />
                        {moment.duration(props.currentHost.hostUptime / 1000 / 1000, 'ms').format("d[d] h[h] m[m] s[s]")}
                        <br />
                        <br />
                    </Col>
                    <Col sm={4} className="text-center">
                        <b>Service Uptime</b>
                        <br />
                        {moment.duration(props.currentHost.serviceUptime / 1000 / 1000, 'ms').format("d[d] h[h] m[m] s[s]")}
                        <br />
                        <br />
                    </Col>
                </Row>
            </Container>
        </Container>
    );
}

function OverviewHostStatus(props) {
    if (props.status === 'success') {
        return <Button size="sm" variant="success" className="status-btn" disabled>{props.name}<br />{Math.round(props.uptime)}%</Button>;
    }
    return <Button size="sm" variant="danger" className="status-btn" disabled>{props.name}<br />{Math.round(props.uptime)}%</Button>;
}

function OverviewUptime(props) {
    var variant = "success";
    if (props.uptime < 90) {
        variant = "warning";
    }
    if (props.uptime < 50) {
        variant = "danger";
    }

    return <Button size="sm" variant={variant} className="status-btn" disabled>{Math.round(props.uptime)}%</Button>;
}

function OverviewHost(props) {
    var statuses = {
        public: {
            http: "error",
            icmp: "error",
            tcp: "error",
            udp: "error"
        },
        internal: {
            http: "error",
            icmp: "error",
            tcp: "error",
            udp: "error"
        }
    }

    if (props.host.latestChecks.public.http) {
        statuses.public.http = props.host.latestChecks.public.http[0].status;
    }
    if (props.host.latestChecks.public.icmp) {
        statuses.public.icmp = props.host.latestChecks.public.icmp[0].status;
    }
    if (props.host.latestChecks.public.tcp) {
        statuses.public.tcp = props.host.latestChecks.public.tcp[0].status;
    }
    if (props.host.latestChecks.public.udp) {
        statuses.public.udp = props.host.latestChecks.public.udp[0].status;
    }

    if (props.host.latestChecks.internal.http) {
        statuses.internal.http = props.host.latestChecks.internal.http[0].status;
    }
    if (props.host.latestChecks.internal.icmp) {
        statuses.internal.icmp = props.host.latestChecks.internal.icmp[0].status;
    }
    if (props.host.latestChecks.internal.tcp) {
        statuses.internal.tcp = props.host.latestChecks.internal.tcp[0].status;
    }
    if (props.host.latestChecks.internal.udp) {
        statuses.internal.udp = props.host.latestChecks.internal.udp[0].status;
    }

    return (
        <tr>
            <td>
                <a
                    className="text-white"
                    href={"http://" + props.host.publicIp}
                >
                    {props.host.hostname}
                </a>
            </td>
            <td>
                <Row>
                    <Col lg={4}>{props.host.publicIp}</Col>
                    <Col lg={8}>
                        <ButtonGroup>
                            <OverviewHostStatus
                                name='HTTP'
                                status={statuses.public.http}
                                uptime={props.host.checkUptime.public.http.percent}
                            />
                            <OverviewHostStatus
                                name='ICMP'
                                status={statuses.public.icmp}
                                uptime={props.host.checkUptime.public.icmp.percent}
                            />
                            <OverviewHostStatus
                                name='TCP'
                                status={statuses.public.tcp}
                                uptime={props.host.checkUptime.public.tcp.percent}
                            />
                            <OverviewHostStatus
                                name='UDP'
                                status={statuses.public.udp}
                                uptime={props.host.checkUptime.public.udp.percent}
                            />
                        </ButtonGroup>
                    </Col>
                </Row>
            </td>
            <td>
                <Row>
                    <Col lg={4}>{props.host.internalIp}</Col>
                    <Col lg={8}>
                        <ButtonGroup>
                            <OverviewHostStatus
                                name='HTTP'
                                status={statuses.internal.http}
                                uptime={props.host.checkUptime.internal.http.percent}
                            />
                            <OverviewHostStatus
                                name='ICMP'
                                status={statuses.internal.icmp}
                                uptime={props.host.checkUptime.internal.icmp.percent}
                            />
                            <OverviewHostStatus
                                name='TCP'
                                status={statuses.internal.tcp}
                                uptime={props.host.checkUptime.internal.tcp.percent}
                            />
                            <OverviewHostStatus
                                name='UDP'
                                status={statuses.internal.udp}
                                uptime={props.host.checkUptime.internal.udp.percent}
                            />
                        </ButtonGroup>
                    </Col>
                </Row>
            </td>
            <td>
                <Link to={"/hosts/" + props.host.id}>details</Link>
            </td>
        </tr>
    );
}

function OverviewHostList(props) {
    return (
        <Container fluid>
            <div className="text-center"><h4>Discovered Hosts</h4></div>
            <Table className="discovered-hosts-table" borderless variant="dark">
                <thead>
                    <tr>
                        <th scope="col">Host</th>
                        <th scope="col">Public</th>
                        <th scope="col">Internal</th>
                        <th scope="col">&nbsp;</th>
                    </tr>
                </thead>
                <tbody>
                    {props.hosts.map((item, key) => {
                        return <OverviewHost host={item} key={key} />;
                    })}
                </tbody>
            </Table>
        </Container>
    );
}

function Overview(props) {
    return (
        <Container fluid>
            <OverviewBanner currentHost={props.currentHost} />
            <br />
            <OverviewHostList hosts={props.hosts} />
        </Container>
    );
}

export default Overview;