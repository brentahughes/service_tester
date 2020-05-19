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
                        <b>Service Restarts</b><br />{props.currentHost.serviceRestarts}<br /><br />
                    </Col>
                    <Col sm={4} className="text-center">
                        <b>First Start</b>
                        <br />
                        <Moment format="YYYY-MM-DD HH:mm:ss" date={props.currentHost.serviceFirstStart} /> UTC
                        <br />
                        <br />
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
                    <Col sm={4} className="text-center">
                        <b>Last Start</b>
                        <br />
                        <Moment format="YYYY-MM-DD HH:mm:ss" date={props.currentHost.serviceLastStart} /> UTC
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
        return <Button size="sm" variant="success" className="status-btn" disabled>{props.name}</Button>;
    }
    return <Button size="sm" variant="danger" className="status-btn" disabled>{props.name}</Button>;
}

function OverviewHost(props) {
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
                            <OverviewHostStatus name='HTTP' status={props.host.latestStatus.public.http} />
                            <OverviewHostStatus name='ICMP' status={props.host.latestStatus.public.icmp} />
                            <OverviewHostStatus name='TCP' status={props.host.latestStatus.public.tcp} />
                            <OverviewHostStatus name='UDP' status={props.host.latestStatus.public.udp} />
                        </ButtonGroup>
                    </Col>
                </Row>
            </td>
            <td>
                <Row>
                    <Col lg={4}>{props.host.internalIp}</Col>
                    <Col lg={8}>
                        <ButtonGroup>
                            <OverviewHostStatus name='HTTP' status={props.host.latestStatus.internal.http} />
                            <OverviewHostStatus name='ICMP' status={props.host.latestStatus.internal.icmp} />
                            <OverviewHostStatus name='TCP' status={props.host.latestStatus.internal.tcp} />
                            <OverviewHostStatus name='UDP' status={props.host.latestStatus.internal.udp} />
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