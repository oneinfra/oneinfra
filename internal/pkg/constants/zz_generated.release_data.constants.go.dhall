let Prelude =
      https://prelude.dhall-lang.org/v20.2.0/package.dhall sha256:a6036bc38d883450598d1de7c98ead113196fe2db02e9733855668b18096f07b

let releaseData/ToJSON =
      (../../../dhall/versionsAsJSON.dhall).releaseData/ToJSON

in  ''
    /**
     * Copyright 2021 Rafael Fernández López <ereslibre@ereslibre.es>
     *
     * Licensed under the Apache License, Version 2.0 (the "License");
     * you may not use this file except in compliance with the License.
     * You may obtain a copy of the License at
     *
     *     http://www.apache.org/licenses/LICENSE-2.0
     *
     * Unless required by applicable law or agreed to in writing, software
     * distributed under the License is distributed on an "AS IS" BASIS,
     * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
     * See the License for the specific language governing permissions and
     * limitations under the License.
     **/

    package constants

    // Code auto-generated. DO NOT EDIT.

    const (
    	// RawReleaseData represents the supported versions for this release
    	RawReleaseData = `${Prelude.JSON.renderYAML releaseData/ToJSON}`
    )
    ''
