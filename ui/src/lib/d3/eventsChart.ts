import * as d3 from 'd3'

import { Event } from 'api/events.type'
import _debounce from 'lodash.debounce'
import day from 'lib/dayjs'
import wrapText from './wrapText'

const margin = {
  top: 15,
  right: 15,
  bottom: 30,
  left: 15,
}

export default function gen({
  root,
  events,
  selectEvent,
}: {
  root: HTMLElement
  events: Event[]
  selectEvent?: (e: Event) => void
}) {
  let width = root.offsetWidth
  const height = root.offsetHeight

  const svg = d3.select(root).append('svg').attr('class', 'chaos-chart').attr('width', width).attr('height', height)

  const halfHourLater = day(events[events.length - 1].start_time).add(0.5, 'h')

  const x = d3
    .scaleLinear()
    .domain([halfHourLater.subtract(1, 'h'), halfHourLater])
    .range([margin.left, width - margin.right])
  const xAxis = d3
    .axisBottom(x)
    .ticks(6)
    .tickFormat(d3.timeFormat('%m-%d %H:%M') as (dv: Date | { valueOf(): number }, i: number) => string)
  const gXAxis = svg
    .append('g')
    .attr('class', 'axis')
    .attr('transform', `translate(0, ${height - margin.bottom})`)
    .call(xAxis)

  // Wrap long text, also used in zoom() and reGen()
  svg.selectAll('.tick text').call(wrapText, 30)

  const allUniqueExperiments = [...new Set(events.map((d) => d.experiment + '/' + d.experiment_id))].map((d) => {
    const [name, uuid] = d.split('/')

    return {
      name,
      uuid,
    }
  })
  const y = d3
    .scaleBand()
    .domain(allUniqueExperiments.map((d) => d.uuid))
    .range([0, height - margin.top - margin.bottom])
    .padding(0.5)
  const yAxis = d3.axisLeft(y).tickFormat('' as any)
  // gYAxis
  svg
    .append('g')
    .attr('class', 'axis')
    .attr('transform', `translate(${margin.left}, ${margin.top})`)
    .call(yAxis)
    .call((g) => g.select('.domain').remove())

  // clipX
  svg
    .append('clipPath')
    .attr('id', 'clip-x-axis')
    .append('rect')
    .attr('x', margin.left)
    .attr('y', 0)
    .attr('width', width - margin.left - margin.right)
    .attr('height', height - margin.bottom)
  const gMain = svg.append('g').attr('clip-path', 'url(#clip-x-axis)')

  const colorPalette = d3
    .scaleOrdinal<string, string>()
    .domain(events.map((d) => d.experiment_id))
    .range(d3.schemeTableau10)

  // legends
  const legendsRoot = d3.select(document.createElement('div')).attr('class', 'chaos-events-legends')
  const legends = legendsRoot
    .selectAll()
    .data(allUniqueExperiments)
    .enter()
    .append('div')
    .on('click', function (d) {
      const _events = events.filter((e) => e.experiment_id === d.uuid)
      const event = _events[_events.length - 1]

      svg
        .transition()
        .duration(750)
        .call(
          zoom.transform as any,
          d3.zoomIdentity
            .translate(width / 2, 0)
            .scale(2)
            .translate(-x(day(event.start_time)), 0)
        )
    })
  legends.append('div').attr('style', (d) => `width: 14px; height: 14px; background: ${colorPalette(d.uuid)};`)
  legends
    .insert('div')
    .attr('style', 'margin-left: 8px; color: rgba(0, 0, 0, 0.54); font-size: 0.75rem; font-weight: bold;')
    .text((d) => d.name)

  function genRectWidth(x: d3.ScaleLinear<number, number>) {
    return (d: Event) => {
      let width = d.finish_time ? x(day(d.finish_time)) - x(day(d.start_time)) : x(day()) - x(day(d.start_time))

      if (width === 0) {
        width = 1
      }

      return width
    }
  }

  const rects = gMain
    .selectAll()
    .data(events)
    .enter()
    .append('rect')
    .attr('x', (d) => x(day(d.start_time)))
    .attr('y', (d) => y(d.experiment_id)! + margin.top)
    .attr('width', genRectWidth(x))
    .attr('height', y.bandwidth())
    .attr('fill', (d) => colorPalette(d.experiment_id))
    .style('cursor', 'pointer')

  const zoom = d3.zoom().scaleExtent([0.1, 5]).on('zoom', zoomed)
  function zoomed() {
    const eventTransform = d3.event.transform

    const newX = eventTransform.rescaleX(x)

    gXAxis.call(xAxis.scale(newX))
    svg.selectAll('.tick text').call(wrapText, 30)
    rects.attr('x', (d) => newX(day(d.start_time))).attr('width', genRectWidth(newX))
  }
  svg.call(zoom as any)

  const tooltip = d3
    .select(document.createElement('div'))
    .attr('class', 'chaos-event-tooltip')
    .call(createTooltip as any)

  function createTooltip(el: d3.Selection<HTMLElement, any, any, any>) {
    el.style('position', 'absolute')
      .style('top', 0)
      .style('left', 0)
      .style('padding', '0.25rem 0.75rem')
      .style('background', '#fff')
      .style('font', '1rem')
      .style('border', '1px solid rgba(0, 0, 0, 0.12)')
      .style('border-radius', '4px')
      .style('opacity', 0)
      .style('transition', 'top 0.25s ease, left 0.25s ease')
      .style('z-index', 999)
  }

  function genTooltipContent(d: Event) {
    return `<b>Experiment: ${d.experiment}</b>
            <br />
            <b>Status: ${d.finish_time ? 'Finished' : 'Running'}</b>
            <br />
            <br />
            <span style="color: rgba(0, 0, 0, 0.67);">Start Time: ${day(d.start_time).format(
              'YYYY-MM-DD HH:mm:ss A'
            )}</span>
            <br />
            ${
              d.finish_time
                ? `<span style="color: rgba(0, 0, 0, 0.67);">Finish Time: ${day(d.finish_time).format(
                    'YYYY-MM-DD HH:mm:ss A'
                  )}</span>`
                : ''
            }
            `
  }

  rects
    .on('click', function (d) {
      if (typeof selectEvent === 'function') {
        selectEvent(d)
      }

      svg
        .transition()
        .duration(750)
        .call(
          zoom.transform as any,
          d3.zoomIdentity
            .translate(width / 2, 0)
            .scale(2)
            .translate(-x(day(d.start_time)), 0)
        )
    })
    .on('mouseover', function (d) {
      let [x, y] = d3.mouse(this)

      tooltip.html(genTooltipContent(d))
      const { width } = tooltip.node()!.getBoundingClientRect()

      y += 50
      if (x > (root.offsetWidth / 3) * 2) {
        x -= width
      }
      if (y > (root.offsetHeight / 3) * 2) {
        y -= 200
      }

      tooltip
        .style('left', x + 'px')
        .style('top', y + 'px')
        .style('opacity', 1)
    })
    .on('mouseleave', function () {
      tooltip.style('opacity', 0)
    })

  function reGen() {
    const newWidth = root.offsetWidth
    width = newWidth

    svg.attr('width', width)
    x.range([margin.left, width - margin.right])
    gXAxis.call(xAxis)
    svg.selectAll('.tick text').call(wrapText, 30)
    rects.attr('x', (d) => x(day(d.start_time))).attr('width', genRectWidth(x))
  }

  d3.select(window).on('resize', _debounce(reGen, 250))

  root.appendChild(legendsRoot.node()!)
  root.appendChild(tooltip.node()!)
  root.style.position = 'relative'
}
