/*
 * Copyright 2022 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { appPath, ignores } from './constants.js'
import { getUIFormEnum, isUIFormIgnore } from './utils.js'

import _get from 'lodash.get'
import _set from 'lodash.set'
import fs from 'fs'
import { nodeToField } from './factory.js'
import sig from 'signale'
import ts from 'typescript'
import yaml from 'js-yaml'

const { factory } = ts

const WARNING_MESSAGE = `/**
 * This file was auto-generated by @ui/openapi.
 * Do not make direct changes to the file.
 */

`

/**
 * Convert the definitions generated by @openapitools/openapi-generator-cli to Formik form data.
 *
 * @param {string} source
 */
export function genForms(source) {
  const chaos = [
    'AWSChaos',
    'DNSChaos',
    'GCPChaos',
    // 'HTTPChaos',
    'IOChaos',
    'JVMChaos',
    'KernelChaos',
    'NetworkChaos',
    'PhysicalMachineChaos',
    'PodChaos',
    'StressChaos',
    'TimeChaos',
  ]

  const program = ts.createProgram([source], {
    target: ts.ScriptTarget.ES2015,
  })
  const sourceFile = program.getSourceFile(source)
  const checker = program.getTypeChecker()
  const nodes = sourceFile.getChildAt(0).getChildren()
  // 1. filter all required schemas
  const interfaces = nodes.filter((node) => node.kind === ts.SyntaxKind.InterfaceDeclaration)

  chaos.forEach((child) => {
    let actions = []
    const objects = []

    // 2. find the corresponding spec
    //
    // all specs will follow the format of `V1alpha1XXXChaosSpec`
    const spec = interfaces.find((node) => node.name.escapedText === `V1alpha1${child}Spec`)

    spec.members.forEach((node) => {
      switch (node.kind) {
        case ts.SyntaxKind.PropertySignature:
          /** @type {string} */
          const identifier = node.name.escapedText || node.name.text
          if (ignores.includes(identifier)) {
            break
          }

          /** @type {string} */
          const comment = node.jsDoc[0].comment ?? '' // prevent error when comment is undefined

          if (identifier === 'action') {
            // get all actions
            actions = getUIFormEnum(comment)
          } else {
            // ignore specifc fields
            if (isUIFormIgnore(comment)) {
              break
            }

            objects.push(nodeToField(identifier, node.type, comment, [], sourceFile, checker))
          }

          break
        default:
          break
      }
    })

    // create data related fields
    const data = factory.createVariableStatement(
      [factory.createModifier(ts.SyntaxKind.ExportKeyword)],
      factory.createVariableDeclarationList(
        [
          factory.createVariableDeclaration(
            factory.createIdentifier('actions'),
            undefined,
            undefined,
            factory.createArrayLiteralExpression(actions.map(factory.createStringLiteral), false)
          ),
          factory.createVariableDeclaration(
            factory.createIdentifier('data'),
            undefined,
            undefined,
            factory.createArrayLiteralExpression(objects, true)
          ),
        ],
        ts.NodeFlags.Const
      )
    )

    const printer = ts.createPrinter({
      omitTrailingSemicolon: true,
    })
    /**
     * Encapsulate printNode method.
     *
     * @param {ts.Node} node
     * @return {string}
     */
    function printNode(node) {
      return printer.printNode(ts.EmitHint.Unspecified, node, sourceFile)
    }
    const dataPrint = printNode(data)

    const file = `${appPath}/src/formik/${child}.ts`
    fs.writeFile(file, WARNING_MESSAGE + dataPrint + '\n', (err) => {
      if (err) {
        sig.error(err)
      } else {
        sig.success(`${child} form data generated`)
      }
    })
  })
}

/**
 * Wrap all specific $refs with allOf to preserve the original siblings.
 *
 * Ref: https://stackoverflow.com/questions/33629750/swagger-schema-properties-ignored-when-using-ref-why
 *
 * @export
 * @param {string} source
 */
export function swaggerRefToAllOf(source) {
  /** @type {object} */
  const swagger = yaml.load(fs.readFileSync(source, 'utf-8'))

  const properties = [
    'definitions["v1alpha1.IOChaosSpec"].properties.attr',
    'definitions["v1alpha1.IOChaosSpec"].properties.mistake',
  ]

  properties.forEach((property) => {
    const p = _get(swagger, property)

    _set(swagger, property, { allOf: [{ $ref: p.$ref }], ...p, $ref: undefined })
  })

  fs.writeFileSync(source, yaml.dump(swagger))
}
